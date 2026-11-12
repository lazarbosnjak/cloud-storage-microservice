package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Postgres struct {
	User     string
	Password string
	DB       string
	Host     string
	Port     string

	MaxConns                   int32
	MinConns                   int32
	MaxConnLifetimeInMinutes   time.Duration
	MaxConnIdleTimeInMinutes   time.Duration
	HealthCheckPeriodInSeconds time.Duration
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

	parseInt := func(key string) (int32, error) {
		value := get(key)
		i64, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("%s: parse %q as int: %w", key, value, err)
		}
		i32 := int32(i64)
		return i32, nil
	}

	maxConns, err := parseInt("MAX_CONNECTIONS")
	if err != nil {
		return Config{}, err
	}
	minConns, err := parseInt("MIN_CONNECTIONS")
	if err != nil {
		return Config{}, err
	}
	maxConnLifetimeInMinutes, err := parseInt("MAX_CONNECTION_LIFETIME_IN_MINUTES")
	if err != nil {
		return Config{}, err
	}
	maxConnIdleTimeInMinutes, err := parseInt("MAX_CONNECTION_IDLE_TIME_IN_MINUTES")
	if err != nil {
		return Config{}, err
	}
	healthCheckPeriodInSeconds, err := parseInt("HEALTH_CHECK_PERIOD_IN_SECONDS")
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Postgres: Postgres{
			User:     get("POSTGRES_USER"),
			Password: get("POSTGRES_PASSWORD"),
			DB:       get("POSTGRES_DB"),
			Host:     get("POSTGRES_HOST"),
			Port:     get("POSTGRES_PORT"),

			MaxConns:                   maxConns,
			MinConns:                   minConns,
			MaxConnLifetimeInMinutes:   time.Minute * time.Duration(maxConnLifetimeInMinutes),
			MaxConnIdleTimeInMinutes:   time.Minute * time.Duration(maxConnIdleTimeInMinutes),
			HealthCheckPeriodInSeconds: time.Second * time.Duration(healthCheckPeriodInSeconds),
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

func (p Postgres) GetDataSourceName() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(p.User, p.Password),
		Host:     fmt.Sprintf("%s:%s", p.Host, p.Port),
		Path:     p.DB,
		RawQuery: "sslmode=disable",
	}
	return u.String()
}
