package auth

import (
	"fmt"
	"os"
	"strings"
)

type Postgres struct {
	User     string
	Password string
	DB       string
	Host     string
	Port     string
}

type Service struct {
	GrpcPort                string
	GracefulShutdownTimeout string
}

type Config struct {
	Postgres Postgres
	Service  Service
}

func Load() (Config, error) {
	var missing []string

	get := func(key string) string {
		v := os.Getenv(key)
		if v == "" {
			missing = append(missing, key)
		}
		return v
	}

	cfg := Config{
		Postgres: Postgres{
			User:     get("POSTGRES_USER"),
			Password: get("POSTGRES_PASSWORD"),
			DB:       get("POSTGRES_DB"),
			Host:     get("POSTGRES_HOST"),
			Port:     get("POSTGRES_PORT"),
		},
		Service: Service{
			GrpcPort:                get("GRPC_PORT"),
			GracefulShutdownTimeout: get("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS"),
		},
	}

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("Missing variables in authentication config: %s", strings.Join(missing, ", "))
	}

	return cfg, nil

}
