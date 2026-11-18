package service

import (
	"auth-service/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
)

type SignInResult struct {
	AccessToken        string
	ExpiresIn          time.Duration
	RefreshToken       string
	MustChangePassword bool
}

func (s *AuthService) SignIn(ctx context.Context, email string, password string) (SignInResult, error) {
	email = normalizeEmail(email)

	if s.validateEmail(email) != nil || s.validatePassword(password) != nil {
		return SignInResult{}, domain.ErrInvalidCredentials
	}

	cred, err := s.store.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return SignInResult{}, fmt.Errorf("get credential: %w", err)
	}

	hash := s.dummyHash
	if cred != nil {
		hash = cred.PasswordHash
	}

	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return SignInResult{}, fmt.Errorf("compare password: %w", err)
	}

	if cred == nil || !match {
		return SignInResult{}, domain.ErrInvalidCredentials
	}

	at, err := s.tokens.IssueAccessToken(cred.ID, cred.UserRole, "")
	if err != nil {
		return SignInResult{}, fmt.Errorf("issue access token: %w", err)
	}

	return SignInResult{
		AccessToken: at.Value,
		ExpiresIn:   at.ExpiresIn,
	}, nil

}
