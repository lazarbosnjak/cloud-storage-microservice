package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	GrpcServerPort          string
	RestServerPort          string
	GracefulShutdownTimeout time.Duration
	UserServiceGRPCAddr     string
}

func NewConfig(logger *slog.Logger) *Config {
	grpcPort := getEnv("GRPC_PORT", logger)
	restPort := getEnv("REST_PORT", logger)
	userServiceGRPCAddr := getEnv("USER_SERVICE_GRPC_ADDR", logger)

	res, err := strconv.Atoi(os.Getenv("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS"))
	if err != nil || res == 0 {
		logger.Error("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS env missing or not a number")
		os.Exit(1)
	}
	gracefulShutdownTimeout := time.Second * time.Duration(res)

	return &Config{
		GrpcServerPort:          grpcPort,
		RestServerPort:          restPort,
		GracefulShutdownTimeout: gracefulShutdownTimeout,
		UserServiceGRPCAddr:     userServiceGRPCAddr,
	}
}

func getEnv(name string, logger *slog.Logger) string {
	env := os.Getenv(name)
	if env == "" {
		logger.Error(fmt.Sprintf("%s env missing", name))
		os.Exit(1)
	}

	return env
}
