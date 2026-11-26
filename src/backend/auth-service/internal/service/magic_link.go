package service

import (
	"auth-service/internal/domain"
	"context"
	"fmt"
	"log/slog"
	"time"
)

func (s *AuthService) RequestMagicLink(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	if s.validateEmail(email) != nil {
		s.wasteTime()
		return nil
	}

	cred, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		slog.Error("magic link lookup", "error", err)
		s.wasteTime()
		return nil
	}

	now := time.Now()
	if !cred.IsEnabled {
		s.wasteTime()
		return nil
	}
	if cred.LockedUntil != nil && cred.LockedUntil.After(now) {
		s.wasteTime()
		return nil
	}

	err = s.verificationTokenStore.DeleteUnusedByUserAndReason(ctx, cred.ID, domain.MagicLink)
	if err != nil {
		return err
	}

	tokenID, err := s.createVerificationToken(ctx, cred.ID, domain.MagicLink, s.config.VerificationToken.ExpiresInFor(domain.MagicLink))
	if err != nil {
		return err
	}

	link := fmt.Sprintf("%s/auth/magic-link/%s", s.config.Service.FrontendAddr, tokenID)
	return s.EmailSender.SendMagicLinkMail([]string{cred.Email}, link)
}

func (s *AuthService) MagicLinkSignIn(ctx context.Context, tokenID string) (SignInResult, error) {
	token, err := s.verificationTokenStore.GetByID(ctx, tokenID)
	if err != nil {
		slog.Warn("magic link sign in: token lookup failed", "token_id", tokenID, "error", err)
		return SignInResult{}, domain.ErrInvalidVerificationToken
	}
	if token.Reason != domain.MagicLink {
		slog.Warn("magic link sign in: wrong token reason", "token_id", tokenID, "error", err)
		return SignInResult{}, domain.ErrInvalidVerificationToken
	}

	if token.UsedAt != nil {
		slog.Warn("magic link sign in: token already used", "token_id", tokenID, "used_at", *token.UsedAt)
		return SignInResult{}, domain.ErrInvalidVerificationToken
	}
	if token.ExpiresAt.Before(time.Now()) {
		slog.Warn("magic link sign in: token expired", "token_id", tokenID, "expires_at", token.ExpiresAt)
		return SignInResult{}, domain.ErrVerificationTokenExpired
	}

	now := time.Now()
	token.UsedAt = &now
	if err := s.verificationTokenStore.Update(ctx, token); err != nil {
		return SignInResult{}, domain.ErrVerificationTokenExpired
	}

	cred, err := s.store.GetByID(ctx, token.UserID)
	if err != nil {
		return SignInResult{}, fmt.Errorf("get credential: %w", err)
	}
	if !cred.IsEnabled {
		return SignInResult{}, domain.ErrInvalidCredentials
	}

	return s.issueSession(ctx, cred)
}
