package main

import (
	"log/slog"
	"os"

	"auth-service/internal/startup"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	startup.Start()

}
