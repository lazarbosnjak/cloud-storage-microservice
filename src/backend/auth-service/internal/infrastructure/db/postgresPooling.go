package db

import (
	"auth-service/internal/startup/config"
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateNewPooling(ctx context.Context, pg config.Postgres) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(pg.GetDataSourceName())
	if err != nil {
		slog.Error("parsing config", "error", err)
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	cfg.MaxConns = pg.MaxConns
	cfg.MinConns = pg.MinConns
	cfg.MaxConnLifetime = pg.MaxConnLifetimeInMinutes
	cfg.MaxConnIdleTime = pg.MaxConnIdleTimeInMinutes
	cfg.HealthCheckPeriod = pg.HealthCheckPeriodInSeconds

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		slog.Error("creating pool", "error", err)
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, pg.PingTimeoutInSeconds)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		slog.Error("ping postgres", "error", err)
		return nil, fmt.Errorf("could not ping postgres: %w", err)
	}

	return pool, nil
}
