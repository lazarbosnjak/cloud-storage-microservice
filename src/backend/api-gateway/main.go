package main

import (
	"log/slog"
	"os"

	"api-gateway/internal/startup"
	"api-gateway/internal/startup/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := config.NewConfig(logger)
	app := startup.NewServer(config, logger)
	app.Start()
}
