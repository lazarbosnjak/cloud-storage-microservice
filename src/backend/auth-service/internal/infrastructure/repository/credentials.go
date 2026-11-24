package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"auth-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CredentialRepository struct {
	pool *pgxpool.Pool
}

func NewCredentialRepository(pool *pgxpool.Pool) *CredentialRepository {
	return &CredentialRepository{pool: pool}
}

func (r *CredentialRepository) Create(ctx context.Context, email string, passwordHash string, userRole domain.Role) (string, error) {
	const query = `
		INSERT INTO credentials (email, password_hash, user_role)
		VALUES ($1, $2, $3)
		RETURNING id`

	var id string

	err := r.pool.QueryRow(ctx, query, email, passwordHash, userRole).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", domain.ErrEmailTaken
		}
		slog.Error("creating credentials", "error", err)
		return "", fmt.Errorf("error while creating credential: %w", err)
	}

	return id, nil
}

func (r *CredentialRepository) GetByEmail(ctx context.Context, email string) (*domain.Credential, error) {
	const query = `
		SELECT id, email, password_hash, user_role, password_changed_at, failed_attempts,
		locked_until, last_failed_at, is_locked, locked_reason, created_at, is_enabled
		FROM credentials
		WHERE email = $1`

	var c domain.Credential
	err := r.pool.QueryRow(ctx, query, email).Scan(&c.ID, &c.Email, &c.PasswordHash, &c.UserRole, &c.PasswordChangedAt, &c.FailedAttempts, &c.LockedUntil, &c.LastFailedAt, &c.IsLocked, &c.LockedReason, &c.CreatedAt, &c.IsEnabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		slog.Error("get email", "error", err)
		return nil, fmt.Errorf("failed to get by email: %w", err)
	}
	return &c, nil
}

func (r *CredentialRepository) UpdateFailedAttempts(ctx context.Context, userID string, failedAttempts uint8, lockedUntil *time.Time) error {
	const query = `
		UPDATE credentials
		SET failed_attempts = $2, last_failed_at = NOW(), locked_until = $3
		WHERE id = $1
	`

	if _, err := r.pool.Exec(ctx, query, userID, failedAttempts, lockedUntil); err != nil {
		slog.Error("update failed attempts", "error", err)
		return fmt.Errorf("error updating failed attempts: %w", err)
	}
	return nil
}

func (r *CredentialRepository) ResetFailedAttempts(ctx context.Context, userID string) error {
	const query = `
		UPDATE credentials
		SET failed_attempts = 0, locked_until = NULL
		WHERE id = $1
	`

	if _, err := r.pool.Exec(ctx, query, userID); err != nil {
		slog.Error("reset failed attempts", "error", err)
		return fmt.Errorf("error resetting failed attempts: %w", err)
	}

	return nil
}

func (r *CredentialRepository) ChangePassword(ctx context.Context, userID string, passwordHash string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin password change: %w", err)
	}

	// Transactional, if one fails, don't write the other and vice versa
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// This query is related to auth credentials
	const query = `
		UPDATE credentials
		SET password_hash = $2,
			password_changed_at = NOW(),
			failed_attempts = 0,
			locked_until = NULL 
			WHERE id = $1
	`

	_, err = tx.Exec(ctx, query, userID, passwordHash)
	if err != nil {
		slog.Error("update password", "error", err)
		return fmt.Errorf("update password: %w", err)
	}

	// This query is related to refresh tokens
	const query2 = `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`

	_, err = tx.Exec(ctx, query2, userID)
	if err != nil {
		slog.Error("revoke refresh tokens", "error", err)
		return fmt.Errorf("revoke refresh tokens: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *CredentialRepository) UpdateIsEnabled(
	ctx context.Context,
	userID string,
	enabled bool,
) error {
	const query = `
		UPDATE credentials
		SET is_enabled = $2
		WHERE id = $1
	`

	if _, err := r.pool.Exec(ctx, query, userID, enabled); err != nil {
		slog.Error("update is_enabled", "error", err)
		return fmt.Errorf("error updating is_enabled: %w", err)
	}
	return nil
}

func (r *CredentialRepository) GetByID(ctx context.Context, userID string) (*domain.Credential, error) {
	const query = `
		SELECT id, email, password_hash, user_role, password_changed_at, failed_attempts,
		locked_until, last_failed_at, is_locked, locked_reason, created_at, is_enabled
		FROM credentials
		WHERE id = $1`

	var c domain.Credential
	err := r.pool.QueryRow(ctx, query, userID).Scan(&c.ID, &c.Email, &c.PasswordHash, &c.UserRole, &c.PasswordChangedAt, &c.FailedAttempts, &c.LockedUntil, &c.LastFailedAt, &c.IsLocked, &c.LockedReason, &c.CreatedAt, &c.IsEnabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		slog.Error("get email", "error", err)
		return nil, fmt.Errorf("failed to get by email: %w", err)
	}
	return &c, nil
}
