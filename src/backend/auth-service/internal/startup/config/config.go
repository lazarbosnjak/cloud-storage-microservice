package config

import (
	"auth-service/internal/domain"
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
	PingTimeoutInSeconds       time.Duration
}

type Argon2id struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint32
	SaltLength  uint32
	KeyLength   uint32
}

type CredentialPolicy struct {
	MaxEmailLength    uint8
	MinPasswordLength uint8
	MaxPasswordLength uint8
	MaxPasswordAge    time.Duration
}

type VerificationToken struct {
	ExpiresIn              time.Duration
	PasswordResetExpiresIn time.Duration
	MagicLinkExpiresIn     time.Duration
}

func (v VerificationToken) ExpiresInFor(reason domain.VerificationTokenReason) time.Duration {
	switch reason {
	case domain.PasswordReset:
		return v.PasswordResetExpiresIn
	case domain.MagicLink:
		return v.MagicLinkExpiresIn
	default:
		return v.ExpiresIn
	}
}

type Service struct {
	GRPCPort                string
	GracefulShutdownTimeout uint32

	AdminEmail    string
	AdminPassword string

	UserServiceGRPCAddr string
	FrontendAddr        string
}

type OTPPolicy struct {
	OtpTTL             time.Duration
	OtpMaxAttempts     uint8
	OtpChallengeLength uint8
}

type Config struct {
	Postgres          Postgres
	Argon2id          Argon2id
	CredentialPolicy  CredentialPolicy
	Service           Service
	SMTP              SMTP
	Token             domain.Token
	Lockout           Lockout
	OTPPolicy         OTPPolicy
	VerificationToken VerificationToken
}

func Load() (Config, error) {
	l := &loader{}

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
			PingTimeoutInSeconds:       time.Second * time.Duration(getInt("PING_TIMEOUT_SECONDS")),
		},
		Argon2id: Argon2id{
			Memory:      getUInt("ARGON2ID_MEMORY_KIBIBYTES"),
			Iterations:  getUInt("ARGON2ID_ITERATIONS"),
			Parallelism: getUInt("ARGON2ID_PARALLELISM"),
			SaltLength:  getUInt("ARGON2ID_SALT_LENGTH"),
			KeyLength:   getUInt("ARGON2ID_KEY_LENGTH"),
		},
		CredentialPolicy: CredentialPolicy{
			MaxEmailLength:    uint8(getUInt("CREDENTIAL_MAX_EMAIL_LENGTH")),
			MinPasswordLength: uint8(getUInt("CREDENTIAL_MIN_PASSWORD_LENGTH")),
			MaxPasswordLength: uint8(getUInt("CREDENTIAL_MAX_PASSWORD_LENGTH")),
			MaxPasswordAge:    l.duration("CREDENTIAL_MAX_PASSWORD_AGE", 60*24*time.Hour),
		},
		Service: Service{
			GRPCPort:                get("GRPC_PORT"),
			GracefulShutdownTimeout: getUInt("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS"),
			AdminEmail:              get("ADMIN_EMAIL"),
			AdminPassword:           get("ADMIN_PASSWORD"),
			UserServiceGRPCAddr:     get("USER_SERVICE_GRPC_ADDR"),
			FrontendAddr:            get("FRONTEND_ADDR"),
		},
		SMTP: SMTP{
			Username: get("SMPT_USERNAME"),
			Password: get("SMPT_PASSWORD"),
			Host:     get("SMPT_HOST"),
			Port:     get("SMPT_PORT"),
		},
		Token:   loadToken(l),
		Lockout: loadLockout(l),
		VerificationToken: VerificationToken{
			ExpiresIn:              l.duration("VERIFICATION_TOKEN_EXPIRATION", time.Minute*15),
			PasswordResetExpiresIn: l.duration("PASSWORD_RESET_TOKEN_EXPIRATION", time.Minute*10),
			MagicLinkExpiresIn:     l.duration("MAGIC_LINK_TOKEN_EXPIRATION", time.Minute*10),
		},
		OTPPolicy: OTPPolicy{
			OtpTTL:             l.duration("OTP_TTL", 5*time.Minute),
			OtpMaxAttempts:     uint8(l.integer("OTP_MAX_ATTEMPTS", 5)),
			OtpChallengeLength: uint8(l.integer("OTP_CHALLENGE_LENGTH", 43)),
		},
	}

	if err := l.err(); err != nil {
		return Config{}, err
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

func getUInt(key string) uint32 {
	return uint32(getInt(key))
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
