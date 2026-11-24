package service

import (
	"context"
	"time"

	"auth-service/internal/domain"
)

type VerificationTokenStore interface {
	Create(ctx context.Context, token *domain.VerificationToken) error
	GetByID(ctx context.Context, id string) (*domain.VerificationToken, error)
	Update(ctx context.Context, token *domain.VerificationToken) error
}

func (s *AuthService) VerifyRegistration(
	ctx context.Context,
	tokenID string,
) error {
	now := time.Now()

	// ***** VALIDATIONS *****

	token, err := s.verificationTokenStore.GetByID(ctx, tokenID)
	if err != nil {
		return err
	}

	reasonRegistration := domain.VerificationTokenReason(0)
	if token.Reason != reasonRegistration {
		return domain.ErrInvalidVerificationToken
	}

	if token.UsedAt != nil {
		// Do nothing, making this seem idempotent
		return nil
	}

	tokenExpired := token.ExpiresAt.Before(now)
	if tokenExpired {
		return domain.ErrVerificationTokenExpired
	}

	// ***** Update token *****

	now = time.Now()
	token.UsedAt = &now
	err = s.verificationTokenStore.Update(ctx, token)
	if err != nil {
		return err
	}

	// ***** Update user credentials *****
	err = s.store.UpdateIsEnabled(ctx, token.UserID, true)
	if err != nil {
		// INFO:
		// Current workaround for transactivity
		// Need to find a way to make token and user updates transactional
		token.UsedAt = nil
		vtsErr := s.verificationTokenStore.Update(ctx, token)
		if vtsErr != nil {
			return vtsErr
		}

		return err
	}

	return nil
}
