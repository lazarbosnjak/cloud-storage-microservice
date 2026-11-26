package config

import "time"

type Config struct {
	GrpcPort                string
	GracefulShutdownTimeout time.Duration

	Cassandra *cassandra
}

func NewConfig() *Config {
	return &Config{
		GrpcPort:                getString("GRPC_PORT"),
		GracefulShutdownTimeout: getDuration("GRACEFUL_SHUTDOWN_TIMEOUT"),
		Cassandra:               newCassandraConfig(),
	}
}
