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
	PingTimeout        time.Duration
	DisconnectTimeout  time.Duration
}

type Config struct {
	GrpcPort                string
	GracefulShutdownTimeout time.Duration
	Mongo                   mongoConfig

	logger *slog.Logger
}

func NewConfig(logger *slog.Logger) *Config {
	grpcPort := getEnv("GRPC_PORT")
	gracefulShutdownTimeout := toTimeDuration(
		getEnv("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS"),
		time.Second)

	mongoInitDBRootUsername := getEnv("MONGO_INITDB_ROOT_USERNAME")
	mongoInitDBRootPassword := getEnv("MONGO_INITDB_ROOT_PASSWORD")
	mongoHost := getEnv("MONGO_HOST")
	mongoPort := getEnv("MONGO_PORT")
	mongoDatabase := getEnv("MONGO_DB")
	mongoPingTimeout := toTimeDuration(
		getEnv("MONGO_PING_TIMEOUT_SECONDS"),
		time.Second)
	mongoDisconnectTimeout := toTimeDuration(
		getEnv("MONGO_DISCONNECT_TIMEOUT_SECONDS"),
		time.Second)

	mongo := mongoConfig{
		Host:               mongoHost,
		Port:               mongoPort,
		Database:           mongoDatabase,
		InitDBRootUsername: mongoInitDBRootUsername,
		InitDBRootPassword: mongoInitDBRootPassword,
		PingTimeout:        mongoPingTimeout,
		DisconnectTimeout:  mongoDisconnectTimeout,
	}

	return &Config{
		GrpcPort:                grpcPort,
		GracefulShutdownTimeout: gracefulShutdownTimeout,
		Mongo:                   mongo,
		logger:                  logger,
	}
}

func getEnv(name string) string {
	env := os.Getenv(name)
	if env == "" {
		slog.Error(fmt.Sprintf("%s env missing", name))
		os.Exit(1)
	}

	return env
}

func toTimeDuration(s string, unit time.Duration) time.Duration {
	res, err := strconv.Atoi(os.Getenv(s))
	if err != nil || res == 0 {
		slog.Error(fmt.Sprintf("%s env missing or not a number", s))
		os.Exit(1)
	}

	return time.Duration(res) * unit
}
