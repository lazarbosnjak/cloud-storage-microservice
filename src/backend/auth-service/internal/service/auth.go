package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"
	"unicode/utf8"

	"auth-service/internal/domain"
	"auth-service/internal/repository"
	"auth-service/internal/startup/config"

	"github.com/alexedwards/argon2id"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidEmail = errors.New("invalid email address")
	ErrPassword     = errors.New("password does not meet requirements")
	ErrEmailTaken   = errors.New("email already taken")
	ErrDefault      = errors.New("there was an error")
)

type CredentialPolicy struct {
	MaxEmailLength    uint8
	MinPasswordLength uint8
	MaxPasswordLength uint8
}

type CredentialStore interface {
	Create(ctx context.Context, email string, passwordHash string, userRole domain.Role) (string, error)
	GetByEmail(ctx context.Context, email string) (*domain.Credential, error)
}

type CreateUserDTO struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
}

type UserClient interface {
	CreateUser(ctx context.Context, user CreateUserDTO) error
}

type AuthService struct {
	store  CredentialStore
	policy CredentialPolicy
	params *argon2id.Params

	UserClient UserClient
}

func (s *AuthService) Ping(ctx context.Context) (string, error) {
	return "hello from auth-service!", nil
}

func NewAuthService(
	store CredentialStore, cfg config.Config,
	userClient UserClient,
) *AuthService {
	return &AuthService{
		store: store,
		policy: CredentialPolicy{
			MaxEmailLength:    cfg.CredentialPolicy.MaxEmailLength,
			MinPasswordLength: cfg.CredentialPolicy.MinPasswordLength,
			MaxPasswordLength: cfg.CredentialPolicy.MaxPasswordLength,
		},
		params: &argon2id.Params{
			Memory:      cfg.Argon2id.Memory * 1024,
			Iterations:  cfg.Argon2id.Iterations,
			Parallelism: uint8(cfg.Argon2id.Parallelism),
			SaltLength:  cfg.Argon2id.SaltLength,
			KeyLength:   cfg.Argon2id.KeyLength,
		},
		UserClient: userClient,
	}
}

func (s *AuthService) Register(ctx context.Context, email string, password string) (string, error) {
	id, err := s.createAccount(ctx, email, password, domain.RoleUser)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidEmail):
			slog.Error("invalid email")
			return "", status.Errorf(codes.InvalidArgument, "invalid email")
		case errors.Is(err, ErrEmailTaken):
			slog.Error("email already registered")
			return "", status.Errorf(codes.AlreadyExists, "email already registered")
		case errors.Is(err, ErrPassword):
			slog.Error("password does not meet the requirements")
			return "", status.Errorf(codes.InvalidArgument, "password does not meet the requirements")
		default:
			slog.Error("register failed", "error", err)
			return "", status.Error(codes.Internal, "internal error")
		}
	}

	// TODO: ADD rpc to user-service
	user := &CreateUserDTO{
		ID:        id,
		Email:     email,
		FirstName: "Lazar",
		LastName:  "Bosnjak",
	}
	s.UserClient.CreateUser(ctx, user)
	return id, nil
}

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
		if errors.Is(err, repository.ErrEmailTaken) {
			slog.Error("email already taken")
			return "", ErrEmailTaken
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
	case !errors.Is(err, repository.ErrNotFound):
		slog.Error("checking admin", "error", err)
		return fmt.Errorf("error while checking admin: %w", err)
	}

	_, err = s.createAccount(ctx, email, password, domain.RoleAdmin)
	if errors.Is(err, repository.ErrEmailTaken) {
		return nil
	}
	return err
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *AuthService) validateEmail(email string) error {
	if len(email) > int(s.policy.MaxEmailLength) {
		return ErrInvalidEmail
	}

	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email {
		return ErrInvalidEmail
	}

	return nil
}

func (s *AuthService) validatePassword(password string) error {
	n := utf8.RuneCountInString(password)
	if n < int(s.policy.MinPasswordLength) || n > int(s.policy.MaxPasswordLength) {
		return ErrPassword
	}

	return nil
}
