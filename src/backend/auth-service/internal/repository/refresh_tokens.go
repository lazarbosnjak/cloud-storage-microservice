package repository

import (
	"auth-service/internal/domain"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{pool: pool}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, rt *domain.RefreshToken) error {
	const query = `
		INSERT INTO refresh_tokens
		(user_id, family_id, token_hash, created_at, expires_at, family_expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query, rt.UserID, rt.FamilyID, rt.TokenHash, rt.CreatedAt, rt.ExpiresAt, rt.FamilyExpiresAt).Scan(&rt.ID)

	return err
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, hash []byte) (*domain.RefreshToken, error) {
	const query = `
		SELECT rt.id, rt.user_id, rt.family_id, rt.token_hash, rt.created_at,
			rt.expires_at, rt.family_expires_at, rt.used_at,
			rt.revoked_at, c.user_role, c.is_enabled
		FROM refresh_tokens rt JOIN credentials c ON c.id = rt.user_id
		WHERE rt.token_hash = $1
	`

	var rt domain.RefreshToken
	err := r.pool.QueryRow(ctx, query, hash).Scan(
		&rt.ID, &rt.UserID, &rt.FamilyID, &rt.TokenHash, &rt.CreatedAt,
		&rt.ExpiresAt, &rt.FamilyExpiresAt, &rt.UsedAt, &rt.RevokedAt,
		&rt.UserRole, &rt.IsEnabled,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		slog.Error("get refresh token", "error", err)
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	return &rt, nil
}

func (r *RefreshTokenRepository) Rotate(ctx context.Context, oldTokenID string, now time.Time, successor *domain.RefreshToken) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin rotate: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	const query1 = `
		UPDATE refresh_tokens SET used_at = $2
		WHERE id = $1 AND used_at IS NULL AND revoked_at IS NULL
	`

	tag, err := tx.Exec(ctx, query1, oldTokenID, now)
	if err == nil && tag.RowsAffected() == 0 {
		return domain.ErrConcurrentRefresh
	}
	if err != nil {
		return fmt.Errorf("consume old token: %w", err)
	}

	const query2 = `
		INSERT INTO refresh_tokens (user_id, family_id, token_hash, created_at, expires_at, family_expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err = tx.QueryRow(ctx, query2, successor.UserID, successor.FamilyID, successor.TokenHash,
		successor.CreatedAt, successor.ExpiresAt, successor.FamilyExpiresAt,
	).Scan(&successor.ID)

	if err != nil {
		return fmt.Errorf("insert successor: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string, at time.Time) error {

	const query = `
		UPDATE refresh_tokens SET revoked_at = $2 WHERE id = $1 AND revoked_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id, at)
	if err != nil {
		slog.Error("revoke refresh token", "error", err)
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepository) RevokeFamily(ctx context.Context, familyID string, at time.Time) error {

	const query = `
		UPDATE refresh_tokens SET revoked_at = $2 WHERE family_id = $1 AND revoked_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, familyID, at)
	if err != nil {
		slog.Error("revoke refresh family", "error", err)
		return fmt.Errorf("revoke refresh family: %w", err)
	}

	return nil
}
