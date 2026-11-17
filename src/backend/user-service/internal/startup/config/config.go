package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type mongoConfig struct {
	Host               string
	Port               string
	Database           string
	InitDBRootUsername string
	InitDBRootPassword string
}

type Config struct {
	GrpcPort                string
	GracefulShutdownTimeout time.Duration
	Mongo                   mongoConfig

	logger *slog.Logger
}

func NewConfig(logger *slog.Logger) *Config {
	grpcPort := getEnv("GRPC_PORT", logger)

	mongoInitDBRootUsername := getEnv("MONGO_INITDB_ROOT_USERNAME", logger)
	mongoInitDBRootPassword := getEnv("MONGO_INITDB_ROOT_PASSWORD", logger)
	mongoHost := getEnv("MONGO_HOST", logger)
	mongoPort := getEnv("MONGO_PORT", logger)
	mongoDatabase := getEnv("MONGO_DB", logger)

	res, err := strconv.Atoi(os.Getenv("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS"))
	if err != nil || res == 0 {
		logger.Error("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS env missing or not a number")
		os.Exit(1)
	}
	gracefulShutdownTimeout := time.Second * time.Duration(res)

	mongo := mongoConfig{
		Host:               mongoHost,
		Port:               mongoPort,
		Database:           mongoDatabase,
		InitDBRootUsername: mongoInitDBRootUsername,
		InitDBRootPassword: mongoInitDBRootPassword,
	}

	return &Config{
		GrpcPort:                grpcPort,
		GracefulShutdownTimeout: gracefulShutdownTimeout,
		Mongo:                   mongo,
		logger:                  logger,
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
