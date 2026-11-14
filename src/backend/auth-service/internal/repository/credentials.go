package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound   = errors.New("credential not fonund")
	ErrEmailTaken = errors.New("email already registered")
	ErrDefault    = errors.New("there was an error")
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

type Credential struct {
	UUID              string
	Email             string
	PasswordHash      string
	UserRole          string
	PasswordChangedAt time.Time
	FailedAttempts    string
	LockedUntil       *time.Time
	EmailVerified     bool
	CreatedAt         time.Time
}

type CredentialRepository struct {
	pool *pgxpool.Pool
}

func NewCredentialRepository(pool *pgxpool.Pool) *CredentialRepository {
	return &CredentialRepository{pool: pool}
}

func (r *CredentialRepository) Create(ctx context.Context, email string, passwordHash string, userRole string) (string, error) {
	const query = `
		INSERT INTO credentials (email, password_hash, user_role)
		VALUES ($1, $2, $3)
		RETURNING uuid`

	var uuid string

	err := r.pool.QueryRow(ctx, query, email, passwordHash, userRole).Scan(&uuid)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", ErrEmailTaken
		}
		return "", fmt.Errorf("error while creating credential: %w", err)
	}

	return uuid, nil
}

func (r *CredentialRepository) GetByEmail(ctx context.Context, email string) (*Credential, error) {
	const query = `
		SELECT uuid, email, password_hash, user_role, password_changed_at, failed_attempts,
		locked_until, email_verified, created_at
		FROM credentials
		WHERE email = $1`

	var c Credential
	err := r.pool.QueryRow(ctx, query, email).Scan(&c.UUID, &c.Email, &c.PasswordHash, &c.UserRole, &c.PasswordChangedAt, &c.FailedAttempts, &c.LockedUntil, &c.EmailVerified, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get by email: %w", err)
	}
	return &c, nil
}
