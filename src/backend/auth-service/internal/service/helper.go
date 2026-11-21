package service

import (
	"auth-service/internal/domain"
	"context"
	"log/slog"
	"math"
	"time"
)

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

func (s *AuthService) recordFailedSignIn(ctx context.Context, cred *domain.Credential, now time.Time) {
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
}

func (s *AuthService) passwordExpired(changedAt time.Time) bool {
	age := s.config.CredentialPolicy.MaxPasswordAge
	return age > 0 && time.Since(changedAt) >= age
}
