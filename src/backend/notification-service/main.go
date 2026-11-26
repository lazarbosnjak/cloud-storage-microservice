package main

import (
	"log/slog"
	"os"

	"notification-service/internal/startup"
	"notification-service/internal/startup/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	cfg := config.NewConfig()
	server := startup.NewServer(cfg, logger)
	server.Start()
}
