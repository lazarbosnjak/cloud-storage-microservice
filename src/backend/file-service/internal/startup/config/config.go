package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	GRPCPort                string
	GracefulShutdownTimeout time.Duration
}

func NewConfig() (*Config, error) {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		return nil, fmt.Errorf("GRPC_PORT is missing")
	}

	timeoutValue := os.Getenv("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS")

	seconds, err := strconv.Atoi(timeoutValue)
	if err != nil || seconds <= 0 {
		return nil, fmt.Errorf(
			"GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS must be a positive integer",
		)
	}

	return &Config{
		GRPCPort:                port,
		GracefulShutdownTimeout: time.Duration(seconds) * time.Second,
	}, nil
}
