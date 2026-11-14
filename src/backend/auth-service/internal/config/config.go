package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
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
	GRPCPort                string
	GracefulShutdownTimeout int32

	AdminEmail    string
	AdminPassword string
}

type Config struct {
	Postgres Postgres
	Service  Service
}

func Load() (Config, error) {

	cfg := Config{
		Postgres: Postgres{
			User:     get("POSTGRES_USER"),
			Password: get("POSTGRES_PASSWORD"),
			DB:       get("POSTGRES_DB"),
			Host:     get("POSTGRES_HOST"),
			Port:     get("POSTGRES_PORT"),

			MaxConns:                   getInt("MAX_CONNECTIONS"),
			MinConns:                   getInt("MIN_CONNECTIONS"),
			MaxConnLifetimeInMinutes:   time.Minute * time.Duration(getInt("MAX_CONNECTION_LIFETIME_IN_MINUTES")),
			MaxConnIdleTimeInMinutes:   time.Minute * time.Duration(getInt("MAX_CONNECTION_IDLE_TIME_IN_MINUTES")),
			HealthCheckPeriodInSeconds: time.Second * time.Duration(getInt("HEALTH_CHECK_PERIOD_IN_SECONDS")),
		},
		Service: Service{
			GRPCPort:                get("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS"),
			GracefulShutdownTimeout: getInt("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS"),
			AdminEmail:              get("ADMIN_EMAIL"),
			AdminPassword:           get("ADMIN_PASSWORD"),
		},
	}

	return cfg, nil

}

func get(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		slog.Warn("reading env", key, "not found")
		return ""
	}

	return val
}

func getInt(key string) int32 {
	val := get(key)
	i32, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		slog.Error("parsing to int", "value", val)
	}
	return int32(i32)
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
