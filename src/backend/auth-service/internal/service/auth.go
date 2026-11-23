package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"auth-service/internal/domain"
	"auth-service/internal/startup/config"
	"auth-service/internal/token"

	"github.com/alexedwards/argon2id"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TokenIssuer interface {
	IssueAccessToken(userID string, role domain.Role, sessionID string) (token.AccessToken, error)
}

type CredentialPolicy struct {
	MaxEmailLength    uint8
	MinPasswordLength uint8
	MaxPasswordLength uint8
}

type CredentialStore interface {
	Create(ctx context.Context, email string, passwordHash string, userRole domain.Role) (string, error)
	GetByEmail(ctx context.Context, email string) (*domain.Credential, error)
	UpdateFailedAttempts(ctx context.Context, userID string, failedAttempts uint8, lockedUntil *time.Time) error
	ResetFailedAttempts(ctx context.Context, userID string) error
	ChangePassword(ctx context.Context, userID string, passwordHash string) error
	UpdateIsEnabled(ctx context.Context, userID string, enabled bool) error
}

type RegisterUserDTO struct {
	Email            string
	Password         string
	RepeatedPassword string
	FirstName        string
	LastName         string
}

type CreateUserDTO struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
}

type UserClient interface {
	CreateUser(ctx context.Context, user *CreateUserDTO) error
}

// TODO: Move to separate file
type EmailSender interface {
	SendVerifyRegistrationMail(to []string, verificationLink string) error
}

// FIXME: move cfg to struct instead of constructor func

type AuthService struct {
	config                 config.Config
	store                  CredentialStore
	refresh                RefreshTokenStore
	verificationTokenStore VerificationTokenStore
	policy                 CredentialPolicy
	params                 *argon2id.Params
	tokens                 TokenIssuer
	dummyHash              string

	refreshIdleTTL     time.Duration
	refreshAbsoluteTTL time.Duration
	refreshReuseGrace  time.Duration

	UserClient  UserClient
	EmailSender EmailSender
}

func (s *AuthService) Ping(ctx context.Context) (string, error) {
	return "hello from auth-service!", nil
}

func NewAuthService(
	store CredentialStore,
	refresh RefreshTokenStore,
	verificationTokenStore VerificationTokenStore,
	cfg config.Config,
	tokens TokenIssuer,
	userClient UserClient,
	emailSender EmailSender,
) (*AuthService, error) {
	params := &argon2id.Params{
		Memory:      cfg.Argon2id.Memory * 1024,
		Iterations:  cfg.Argon2id.Iterations,
		Parallelism: uint8(cfg.Argon2id.Parallelism),
		SaltLength:  cfg.Argon2id.SaltLength,
		KeyLength:   cfg.Argon2id.KeyLength,
	}

	// Dummy hash is used to "waste time" in order to prevent time based hacks
	dummy, err := argon2id.CreateHash("pouch-dummy-password", params)
	if err != nil {
		slog.Error("create dummy hash", "error", err)
		return nil, fmt.Errorf("create dummy hash: %w", err)
	}

	return &AuthService{
		config:                 cfg,
		store:                  store,
		refresh:                refresh,
		verificationTokenStore: verificationTokenStore,
		policy: CredentialPolicy{
			MaxEmailLength:    cfg.CredentialPolicy.MaxEmailLength,
			MinPasswordLength: cfg.CredentialPolicy.MinPasswordLength,
			MaxPasswordLength: cfg.CredentialPolicy.MaxPasswordLength,
		},
		params:             params,
		tokens:             tokens,
		dummyHash:          dummy,
		refreshIdleTTL:     cfg.Token.RefreshIdleTTL,
		refreshAbsoluteTTL: cfg.Token.RefreshAbsoluteTTL,
		refreshReuseGrace:  cfg.Token.RefreshReuseGrace,
		UserClient:         userClient,
		EmailSender:        emailSender,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, user RegisterUserDTO) (string, error) {
	if user.RepeatedPassword != user.Password {
		return "", status.Error(codes.InvalidArgument, domain.ErrPasswordMismatch.Error())
	}

	id, err := s.createAccount(ctx, user.Email, user.Password, domain.RoleUser)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidEmail):
			slog.Error("invalid email")
			return "", status.Errorf(codes.InvalidArgument, "invalid email")
		case errors.Is(err, domain.ErrEmailTaken):
			slog.Error("email already registered")
			return "", status.Errorf(codes.AlreadyExists, "email already registered")
		case errors.Is(err, domain.ErrPassword):
			slog.Error("password does not meet the requirements")
			return "", status.Errorf(codes.InvalidArgument, "password does not meet the requirements")
		default:
			slog.Error("register failed", "error", err)
			return "", status.Error(codes.Internal, "internal error")
		}
	}

	createUserDTO := &CreateUserDTO{
		ID:        id,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
	err = s.UserClient.CreateUser(ctx, createUserDTO)
	if err != nil {
		slog.Error("error creating user", "err", err)
		return "", status.Error(codes.Unavailable, "error creating user")
	}

	link, err := s.constructRegistrationVerificationLink(ctx, id)
	if err != nil {
		return "", err
	}

	err = s.EmailSender.SendVerifyRegistrationMail([]string{user.Email}, link)
	if err != nil {
		return "", err
	}

	return id, nil
}

// TODO:
// - pass RegisterUserDTO instead of email and password
// - move repeated password matching to createAccount()

func (s *AuthService) createAccount(ctx context.Context, email string, password string, userRole domain.Role) (string, error) {
	email = normalizeEmail(email)
	if err := s.validateEmail(email); err != nil {
		return "", err
	}
	if err := s.validatePassword(password); err != nil {
		return "", err
	}

	hash, err := argon2id.CreateHash(password, s.params)
	if err != nil {
		slog.Error("hashing password", "error", err)
		return "", fmt.Errorf("error while hashing password: %w", err)
	}

	id, err := s.store.Create(ctx, email, hash, userRole)
	if err != nil {
		if errors.Is(err, domain.ErrEmailTaken) {
			slog.Error("email already taken")
			return "", domain.ErrEmailTaken
		}
		slog.Error("creading credential", "error", err)
		return "", fmt.Errorf("error while creating credential: %w", err)
	}

	return id, nil
}

func (s *AuthService) EnsureAdmin(ctx context.Context, email string, password string) error {
	email = normalizeEmail(email)

	existing, err := s.store.GetByEmail(ctx, email)
	switch {
	case err == nil:
		if existing.UserRole != domain.RoleAdmin {
			slog.Warn("admin bootstrap skipped: account exists with a different role")
		}
		return nil
	case !errors.Is(err, domain.ErrNotFound):
		slog.Error("checking admin", "error", err)
		return fmt.Errorf("error while checking admin: %w", err)
	}

	_, err = s.createAccount(ctx, email, password, domain.RoleAdmin)
	if errors.Is(err, domain.ErrEmailTaken) {
		return nil
	}
	return err
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *AuthService) validateEmail(email string) error {
	if len(email) > int(s.policy.MaxEmailLength) {
		return domain.ErrInvalidEmail
	}

	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email {
		return domain.ErrInvalidEmail
	}

	return nil
}

func (s *AuthService) validatePassword(password string) error {
	n := utf8.RuneCountInString(password)
	if n < int(s.policy.MinPasswordLength) || n > int(s.policy.MaxPasswordLength) {
		return domain.ErrPassword
	}

	return nil
}

func (s *AuthService) createVerificationToken(
	ctx context.Context,
	userID string,
	reason domain.VerificationTokenReason,
) (string, error) {
	token := &domain.VerificationToken{
		ID:        "",
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(s.config.VerificationToken.ExpiresIn),
		UsedAt:    nil,
		Reason:    reason,
	}

	err := s.verificationTokenStore.Create(ctx, token)
	if err != nil {
		return "", err
	}

	return token.ID, nil
}

func (s *AuthService) constructRegistrationVerificationLink(
	ctx context.Context,
	userID string,
) (string, error) {
	token, err := s.createVerificationToken(ctx, userID, domain.VerificationTokenReason(0))
	if err != nil {
		return "", err
	}

	frontendAddr := s.config.Service.FrontendAddr
	path := "/api/auth/register/verify"
	link := fmt.Sprintf(
		"%s%s/%s",
		frontendAddr,
		path,
		token,
	)

	return link, nil
}
