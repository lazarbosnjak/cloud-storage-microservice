package main

import (
	"log/slog"
	"os"

	"file-service/internal/startup"
	"file-service/internal/startup/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Error("invalid configuration", "err", err)
		os.Exit(1)
	}

	app := startup.NewServer(cfg, logger)

	if err := app.Start(); err != nil {
		logger.Error("service exited with error", "err", err)
		os.Exit(1)
	}
}
