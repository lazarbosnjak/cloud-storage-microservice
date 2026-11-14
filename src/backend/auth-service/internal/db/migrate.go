package db

import (
	"auth-service/migrations"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func RunMigrations(dsn string) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("error while loading migrations: %w", err)
	}

	url := "pgx5://" + strings.TrimPrefix(dsn, "postgres://")

	m, err := migrate.NewWithSourceInstance("iofs", src, url)
	if err != nil {
		return fmt.Errorf("failed to init migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("database schema up to date")
			return nil
		}
		return fmt.Errorf("error while aplying migrations: %w", err)
	}

	slog.Info("database migrations applied")
	return nil
}
