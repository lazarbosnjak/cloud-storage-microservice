package service

import (
	"auth-service/internal/domain"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
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

	if cred != nil && cred.LockedUntil != nil && cred.LockedUntil.After(now) {
		// Uncoment line below for testing purposes. We avoid sending this info since that would tell hacker that account exists
		// return SignInResult{}, domain.ErrAccountLocked
		return SignInResult{}, domain.ErrInvalidCredentials
	}

	// dummyHash is just holding random hash in case there are no users registered. Why? Because it "wastes time" preventing time driven attacks
	hash := s.dummyHash
	if cred != nil {
		hash = cred.PasswordHash
	}

	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return SignInResult{}, fmt.Errorf("compare password: %w", err)
	}

	if cred == nil {
		return SignInResult{}, domain.ErrInvalidCredentials
	}

	if !match {
		streak := s.nextStreak(cred.FailedAttempts, cred.LastFailedAt, now)

		var lockedUntil *time.Time
		if d, locked := s.lockoutFor(streak); locked {
			deadline := now.Add(d)
			lockedUntil = &deadline

			slog.Warn("account locked", "user_id", cred.ID, "streak", streak, "locked_until", deadline)
		}

		if err := s.store.UpdateFailedAttempts(ctx, cred.ID, streak, lockedUntil); err != nil {
			slog.Error("update failed attempts", "error", err)
		}
		return SignInResult{}, domain.ErrInvalidCredentials
	}

	if !cred.IsEnabled {
		return SignInResult{}, domain.ErrInvalidCredentials
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

func (s *AuthService) nextStreak(prev uint8, LastFailedAt *time.Time, now time.Time) uint8 {
	if prev == math.MaxUint8 {
		return math.MaxUint8
	}
	if LastFailedAt == nil || now.Sub(*LastFailedAt) >= s.config.Lockout.ResetWindow {
		return 1
	}
	return prev + 1
}

func (s *AuthService) lockoutFor(streak uint8) (time.Duration, bool) {
	treshold := uint8(s.config.Lockout.Treshold)
	if streak < treshold {
		return 0, false
	}

	d := s.config.Lockout.BaseDuration
	for i := int(streak - treshold); i > 0 && d < s.config.Lockout.MaxDuration; i-- {
		d *= 2
	}
	if d > s.config.Lockout.MaxDuration {
		d = s.config.Lockout.MaxDuration
	}
	return d, true
}
