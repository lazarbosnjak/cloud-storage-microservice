package service

import (
	"context"

	"auth-service/internal/domain"
)

type VerificationTokenStore interface {
	Create(ctx context.Context, token *domain.VerificationToken) error
}
