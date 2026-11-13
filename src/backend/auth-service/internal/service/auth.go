package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"runtime"
	"strings"
	"unicode/utf8"

	"auth-service/internal/repository"

	"github.com/alexedwards/argon2id"
)

var (
	ErrInvalidEmail = errors.New("invalid email address")
	ErrPassword     = errors.New("password does not meet requirements")
	ErrEmailTaken   = errors.New("email already taken")
	ErrDefault      = errors.New("there was an error")
)

const (
	maxEmailLen    = 254
	minPasswordLen = 8
	maxPasswordLen = 128
)

type CredentialStore interface {
	Create(ctx context.Context, email string, passwordHash string) (string, error)
	GetByEmail(ctx context.Context, email string) (*repository.Credential, error)
}

type AuthService struct {
	store  CredentialStore
	params *argon2id.Params
}

func NewAuthService(store CredentialStore) *AuthService {
	return &AuthService{
		store: store,
		params: &argon2id.Params{
			Memory:      128 * 1024,
			Iterations:  4,
			Parallelism: uint8(runtime.NumCPU()),
			SaltLength:  16,
			KeyLength:   32,
		},
	}
}

func (s *AuthService) Register(ctx context.Context, email string, password string) (string, error) {
	email = normalizeEmail(email)
	if err := validateEmail(email); err != nil {
		return "", err
	}
	if err := validatePassword(password); err != nil {
		return "", err
	}

	hash, err := argon2id.CreateHash(password, s.params)
	if err != nil {
		return "", fmt.Errorf("error while hashing password: %w", err)
	}

	id, err := s.store.Create(ctx, email, hash)
	if err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			return "", ErrEmailTaken
		}
		return "", fmt.Errorf("error while creating credential: %w", err)
	}

	return id, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateEmail(email string) error {
	if len(email) > maxEmailLen {
		return ErrInvalidEmail
	}

	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email {
		return ErrInvalidEmail
	}

	return nil
}

func validatePassword(password string) error {
	n := utf8.RuneCountInString(password)
	if n < minPasswordLen || n > maxPasswordLen {
		return ErrPassword
	}

	return nil
}
