package db

import (
	"auth-service/internal/config"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateNewPooling(ctx context.Context, c config.Postgres) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(c.GetDataSourceName())
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	cfg.MaxConns = c.MaxConns
	cfg.MinConns = c.MinConns
	cfg.MaxConnLifetime = c.MaxConnLifetimeInMinutes
	cfg.MaxConnIdleTime = c.MaxConnIdleTimeInMinutes
	cfg.HealthCheckPeriod = c.HealthCheckPeriodInSeconds

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("could not ping postgres: %w", err)
	}

	return pool, nil
}
