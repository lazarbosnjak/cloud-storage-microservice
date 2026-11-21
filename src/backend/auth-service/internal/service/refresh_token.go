package service

import (
	"auth-service/internal/domain"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	cryptorand "crypto/rand"
	"crypto/sha256"
)

type RefreshTokenStore interface {
	Create(ctx context.Context, rt *domain.RefreshToken) error
	GetByHash(ctx context.Context, hash []byte) (*domain.RefreshToken, error)
	Rotate(ctx context.Context, oldTokenID string, now time.Time, successor *domain.RefreshToken) error
	Revoke(ctx context.Context, id string, at time.Time) error
	RevokeFamily(ctx context.Context, familyID string, at time.Time) error
}

type RefreshResult struct {
	AccessToken  string
	ExpiresIn    time.Duration
	RefreshToken string
}

func generateOpaqueToken() (string, []byte, error) {
	raw := cryptorand.Text()
	return raw, hashToken(raw), nil
}

func hashToken(raw string) []byte {
	h := sha256.Sum256([]byte(raw))
	return h[:]
}

func (s *AuthService) Refresh(ctx context.Context, raw string) (RefreshResult, error) {
	if raw == "" || len(raw) > 128 {
		return RefreshResult{}, domain.ErrInvalidToken
	}

	rt, err := s.refresh.GetByHash(ctx, hashToken(raw))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return RefreshResult{}, domain.ErrInvalidToken
		}
		return RefreshResult{}, fmt.Errorf("get refresh token: %w", err)
	}

	now := time.Now()

	if rt.UsedAt != nil {
		if now.Sub(*rt.UsedAt) > s.refreshReuseGrace {
			if err := s.refresh.RevokeFamily(ctx, rt.FamilyID, now); err != nil {
				slog.Error("revoke family on reuse", "error", err)
			}
			slog.Warn("refresh token reuse detected, family revoked", "user_id", rt.UserID, "family_id", rt.FamilyID)
		}
		return RefreshResult{}, domain.ErrInvalidToken
	}

	if rt.RevokedAt != nil {
		if err := s.refresh.RevokeFamily(ctx, rt.FamilyID, now); err != nil {
			slog.Error("revoke family on revoke replay", "error", err)
		}
		return RefreshResult{}, domain.ErrInvalidToken
	}

	if !rt.IsEnabled {
		return RefreshResult{}, domain.ErrInvalidToken
	}
	if !now.Before(rt.ExpiresAt) {
		return RefreshResult{}, domain.ErrInvalidToken
	}
	if !now.Before(rt.FamilyExpiresAt) {
		return RefreshResult{}, domain.ErrInvalidToken
	}

	rawNext, hashNext, err := generateOpaqueToken()
	if err != nil {
		return RefreshResult{}, fmt.Errorf("generate successor token: %w", err)
	}

	successor := domain.RefreshToken{
		UserID:    rt.UserID,
		FamilyID:  rt.FamilyID,
		TokenHash: hashNext,
		CreatedAt: now,
	}
	successor.ExpiresAt = now.Add(s.refreshIdleTTL)

	if successor.ExpiresAt.After(rt.FamilyExpiresAt) {
		successor.ExpiresAt = rt.FamilyExpiresAt
	}
	successor.FamilyExpiresAt = rt.FamilyExpiresAt

	if err := s.refresh.Rotate(ctx, rt.ID, now, &successor); err != nil {
		if errors.Is(err, domain.ErrConcurrentRefresh) {
			return RefreshResult{}, domain.ErrInvalidToken
		}
		return RefreshResult{}, fmt.Errorf("rotate refresh token: %w", err)
	}

	at, err := s.tokens.IssueAccessToken(rt.UserID, rt.UserRole, rt.FamilyID)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("issue access token: %w", err)
	}

	return RefreshResult{AccessToken: at.Value, ExpiresIn: at.ExpiresIn, RefreshToken: rawNext}, nil
}

func (s *AuthService) SignOut(ctx context.Context, raw string) error {
	if raw == "" {
		return nil
	}
	rt, err := s.refresh.GetByHash(ctx, hashToken(raw))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("get refresh token: %w", err)
	}
	if err := s.refresh.RevokeFamily(ctx, rt.FamilyID, time.Now()); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}
