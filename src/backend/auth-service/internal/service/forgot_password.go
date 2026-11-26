package service

import (
	"auth-service/internal/domain"
	"context"
	"fmt"
	"log/slog"
)

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	if s.validateEmail(email) != nil {
		return nil // you can also return 'domain.ErrInvalidCredentials', but there is no point since client won't see this message
	}

	cred, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		slog.Error("forgot password lookup", "error", err)
		return nil
	}

	tokenID, err := s.createVerificationToken(ctx, cred.ID, domain.PasswordReset)
	if err != nil {
		return err
	}

	link := fmt.Sprintf("%s/password/reset/%s", s.config.Service.FrontendAddr, tokenID)
	return s.EmailSender.SendPasswordResetMail([]string{cred.Email}, link)

}
