package repository

import (
	"auth-service/internal/domain"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OtpCodeRepository struct {
	pool *pgxpool.Pool
}

func NewOtpCodeRepository(pool *pgxpool.Pool) *OtpCodeRepository {
	return &OtpCodeRepository{pool: pool}
}

func (r *OtpCodeRepository) Create(ctx context.Context, otp *domain.OtpCode) error {
	const query = `
		INSERT INTO otp_codes (user_id, code_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, query, otp.UserID, otp.CodeHash, otp.ExpiresAt).Scan(&otp.ID)

	return err
}

func (r *OtpCodeRepository) InvalidateAllForUser(ctx context.Context, userID string) error {
	const query = `
		UPDATE otp_codes SET used_at = NOW() 
		WHERE user_id = $1 AND used_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *OtpCodeRepository) GetByID(ctx context.Context, id string) (*domain.OtpCode, error) {
	const query = `
		SELECT id, user_id, code_hash, attempts, created_at, expires_at, used_at
		FROM otp_codes
		WHERE id = $1
	`

	var otp domain.OtpCode
	err := r.pool.QueryRow(ctx, query, id).Scan(&otp.ID, &otp.UserID, &otp.CodeHash, &otp.Attempts, &otp.CreatedAt, &otp.ExpiresAt, &otp.UsedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		slog.Error("get otp code", "error", err)
		return nil, fmt.Errorf("get otp code: %w", err)
	}

	return &otp, nil

}

func (r *OtpCodeRepository) Consume(ctx context.Context, id string) error {
	const query = `
		UPDATE otp_codes SET used_at = NOW() 
		WHERE id = $1 AND used_at IS NULL
	`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("consume otp code: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConcurrentRefresh
	}

	return nil

}

func (r *OtpCodeRepository) IncrementAttempts(ctx context.Context, id string, attempts int) error {
	const query = `
		UPDATE otp_codes SET attempts = $2
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *OtpCodeRepository) DeleteExpired(ctx context.Context) (int64, error) {
	const query = `
		DELETE FROM otp_codes
		WHERE expires_at < NOW()
	`
	tag, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("delete expired otp codes: %w", err)
	}

	return tag.RowsAffected(), nil
}
