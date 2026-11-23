package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"auth-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("error while creating verification token: %w", err)
		}
		slog.Error("creating verification token", "error", err)
		return fmt.Errorf("error while creating verification token: %w", err)
	}

	return nil
}

func (r *VerificationTokenRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.VerificationToken, error) {
	const query = `
		SELECT id, user_id, created_at, expires_at, used_at, reason
		FROM verification_token
		WHERE id = $1;
	`

	var token domain.VerificationToken
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&token.ID, &token.UserID,
		&token.CreatedAt, &token.ExpiresAt, &token.UsedAt,
		&token.Reason)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVerificationTokenNotFound
		}
		slog.Error("get verification token", "error", err)
		return nil, fmt.Errorf("failed to get verification token by id: %w", err)
	}

	return &token, err
}

func (r *VerificationTokenRepository) Update(
	ctx context.Context,
	token *domain.VerificationToken,
) error {
	const query = `
		UPDATE verification_token
		SET 
		user_id = $2,
		created_at = $3,
		expires_at = $4,
		used_at = $5,
		reason = $6
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		token.ID,
		token.UserID,
		token.CreatedAt,
		token.ExpiresAt,
		token.UsedAt,
		token.Reason)
	if err != nil {
		slog.Error("update verification token", "error", err)
		return fmt.Errorf("error updating verification token: %w", err)
	}

	return nil
}
