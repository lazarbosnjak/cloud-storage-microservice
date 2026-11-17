package main

import (
	"log/slog"
	"os"

	"user-service/internal/startup"
	"user-service/internal/startup/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	config := config.NewConfig(logger)
	app := startup.NewServer(config, logger)
	app.Start()
}
