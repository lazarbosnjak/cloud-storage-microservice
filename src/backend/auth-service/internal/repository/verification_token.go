package repository

import (
	"context"

	"auth-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type VerificationTokenRepository struct {
	pool *pgxpool.Pool
}

func NewVerificationTokenRepository(pool *pgxpool.Pool) *VerificationTokenRepository {
	return &VerificationTokenRepository{pool: pool}
}

func (r *VerificationTokenRepository) Create(
	ctx context.Context,
	token *domain.VerificationToken,
) error {
	const query = `
		INSERT INTO verification_token
		(user_id, created_at, expires_at, reason)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query,
		token.UserID,
		token.CreatedAt,
		token.ExpiresAt,
		token.Reason).Scan(&token.ID)

	return err
}
