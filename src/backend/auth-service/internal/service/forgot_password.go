package service

import (
	"auth-service/internal/domain"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/alexedwards/argon2id"
)

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	if s.validateEmail(email) != nil {
		s.wasteTime()
		return nil // you can also return 'domain.ErrInvalidCredentials', but there is no point since client won't see this message
	}

	cred, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		slog.Error("forgot password lookup", "error", err)
		s.wasteTime()
		return nil
	}

	err = s.verificationTokenStore.DeleteUnusedByUserAndReason(ctx, cred.ID, domain.PasswordReset)
	if err != nil {
		return err
	}

	tokenID, err := s.createVerificationToken(ctx, cred.ID, domain.PasswordReset, s.config.VerificationToken.ExpiresInFor(domain.PasswordReset))
	if err != nil {
		return err
	}

	link := fmt.Sprintf("%s/password/reset/%s", s.config.Service.FrontendAddr, tokenID)
	return s.EmailSender.SendPasswordResetMail([]string{cred.Email}, link)
}

func (s *AuthService) wasteTime() {
	_, _ = argon2id.ComparePasswordAndHash("pouch-dummy-password", s.dummyHash)
}

func (s *AuthService) ResetPassword(ctx context.Context, tokenID, newPassword, repeatedPassword string) error {
	if newPassword != repeatedPassword {
		return domain.ErrPasswordMismatch
	}

	if err := s.validatePassword(newPassword); err != nil {
		return domain.ErrPassword
	}

	token, err := s.verificationTokenStore.GetByID(ctx, tokenID)
	if err != nil {
		return domain.ErrVerificationTokenNotFound
	}
	if token.Reason != domain.PasswordReset {
		return domain.ErrInvalidVerificationToken
	}
	if token.UsedAt != nil {
		return domain.ErrInvalidVerificationToken
	}
	if token.ExpiresAt.Before(time.Now()) {
		return domain.ErrVerificationTokenExpired
	}

	hash, err := argon2id.CreateHash(newPassword, s.params)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	if err := s.store.ChangePassword(ctx, token.UserID, hash); err != nil {
		return err
	}

	now := time.Now()
	token.UsedAt = &now
	return s.verificationTokenStore.Update(ctx, token)
}
