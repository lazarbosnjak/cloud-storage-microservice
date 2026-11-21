package service

import (
	"auth-service/internal/domain"
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
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

	now := time.Now()

	// dummyHash is just holding random hash in case there are no users registered. Why? Because it "wastes time" preventing time driven attacks
	hash := s.dummyHash
	if cred != nil {
		hash = cred.PasswordHash
	}

	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return SignInResult{}, fmt.Errorf("compare password: %w", err)
	}

	if cred != nil && cred.LockedUntil != nil && cred.LockedUntil.After(now) {
		return SignInResult{}, domain.ErrAccountLocked
	}

	if cred == nil {
		return SignInResult{}, domain.ErrInvalidCredentials
	}

	if !match {
		s.recordFailedSignIn(ctx, cred, now)
		return SignInResult{}, domain.ErrInvalidCredentials
	}

	if !cred.IsEnabled {
		return SignInResult{}, domain.ErrInvalidCredentials
	}

	if s.passwordExpired(cred.PasswordChangedAt) {
		return SignInResult{MustChangePassword: true}, nil
	}

	if cred.FailedAttempts > 0 || cred.LockedUntil != nil {
		if err := s.store.ResetFailedAttempts(ctx, cred.ID); err != nil {
			slog.Error("reset failed attempts", "error", err)
		}
	}

	at, err := s.tokens.IssueAccessToken(cred.ID, cred.UserRole, "") // Currently empty session id is passed since we haven't implemented refresh tokens yet.
	if err != nil {
		return SignInResult{}, fmt.Errorf("issue access token: %w", err)
	}

	return SignInResult{
		AccessToken: at.Value,
		ExpiresIn:   at.ExpiresIn,
	}, nil

}

func (s *AuthService) ChangePassword(ctx context.Context, email string, currentPassword string, newPassword string) error {
	email = normalizeEmail(email)

	if s.validateEmail(email) != nil {
		return domain.ErrInvalidCredentials
	}

	if err := s.validatePassword(newPassword); err != nil {
		return domain.ErrPassword
	}

	cred, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		return domain.ErrInvalidCredentials
	}

	if cred.LockedUntil != nil && cred.LockedUntil.After(time.Now()) {
		return domain.ErrAccountLocked
	}

	match, err := argon2id.ComparePasswordAndHash(currentPassword, cred.PasswordHash)
	if err != nil {
		slog.Error("compare password", "error", err)
		return fmt.Errorf("compare password: %w", err)
	}
	if !match {
		s.recordFailedSignIn(ctx, cred, time.Now())
		return domain.ErrInvalidCredentials
	}

	hash, err := argon2id.CreateHash(newPassword, s.params)
	if err != nil {
		slog.Error("hash new password", "error", err)
		return fmt.Errorf("hash new password: %w", err)
	}

	// Making sure that old password is not same as new one
	if subtle.ConstantTimeCompare([]byte(newPassword), []byte(currentPassword)) == 1 {
		return domain.ErrInvalidCredentials
	}

	if err := s.store.ChangePassword(ctx, cred.ID, hash); err != nil {
		slog.Error("change password", "error", err)
		return fmt.Errorf("change password: %w", err)
	}

	return nil
}
