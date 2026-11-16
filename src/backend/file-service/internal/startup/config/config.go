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

	MongoHost              string
	MongoPort              string
	MongoUsername          string
	MongoPassword          string
	MongoDB                string
	MongoPingTimeout       time.Duration
	MongoDisconnectTimeout time.Duration

	PingDelay time.Duration
}

func NewConfig() (*Config, error) {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		return nil, fmt.Errorf("GRPC_PORT is missing")
	}

	gracefulTimeout, err := readPositiveDuration(
		"GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS",
	)
	if err != nil {
		return nil, err
	}

	mongoHost, err := readRequired("MONGO_HOST")
	if err != nil {
		return nil, err
	}

	mongoPort, err := readRequired("MONGO_PORT")
	if err != nil {
		return nil, err
	}

	mongoUsername, err := readRequired("MONGO_INITDB_ROOT_USERNAME")
	if err != nil {
		return nil, err
	}

	mongoPassword, err := readRequired("MONGO_INITDB_ROOT_PASSWORD")
	if err != nil {
		return nil, err
	}

	mongoDB, err := readRequired("MONGO_DB")
	if err != nil {
		return nil, err
	}

	mongoPingTimeout, err := readPositiveDuration(
		"MONGO_PING_TIMEOUT_SECONDS",
	)
	if err != nil {
		return nil, err
	}

	mongoDisconnectTimeout, err := readPositiveDuration(
		"MONGO_DISCONNECT_TIMEOUT_SECONDS",
	)
	if err != nil {
		return nil, err
	}

	pingDelay, err := readPositiveDuration(
		"PING_DELAY_SECONDS",
	)
	if err != nil {
		return nil, err
	}

	return &Config{
		GRPCPort:                port,
		GracefulShutdownTimeout: gracefulTimeout,

		MongoHost:              mongoHost,
		MongoPort:              mongoPort,
		MongoUsername:          mongoUsername,
		MongoPassword:          mongoPassword,
		MongoDB:                mongoDB,
		MongoPingTimeout:       mongoPingTimeout,
		MongoDisconnectTimeout: mongoDisconnectTimeout,

		PingDelay: pingDelay,
	}, nil
}

// proverava da li obavezna environment promenljiva postoji
func readRequired(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%s is missing", key)
	}
	return value, nil
}

// ucitava broj sekundi i pretvara ga u time.Duration
func readPositiveDuration(key string) (time.Duration, error) {
	value := os.Getenv(key)

	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}

	return time.Duration(seconds) * time.Second, nil
}
