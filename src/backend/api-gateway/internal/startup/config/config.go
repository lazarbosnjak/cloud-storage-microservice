package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	GrpcServerPort              string
	RestServerPort              string
	GracefulShutdownTimeout     time.Duration
	UserServiceGRPCAddr         string
	AuthServiceGRPCAddr         string
	JwtPublicKeyPath            string
	FileServiceGRPCAddr         string
	NotificaitonServiceGRPCAddr string

	RateLimitRPS        uint
	RateLimitBucketSize uint
	RateLimitIdleTTL    time.Duration
}

func NewConfig(logger *slog.Logger) *Config {
	grpcPort := getString("GRPC_PORT")
	restPort := getString("REST_PORT")
	userServiceGRPCAddr := getString("USER_SERVICE_GRPC_ADDR")
	authServiceGRPCAddr := getString("AUTH_SERVICE_GRPC_ADDR")
	jwtPublicKeyPath := getString("JWT_PUBLIC_KEY_PATH")
	fileServiceGRPCAddr := getString("FILE_SERVICE_GRPC_ADDR")
	notifServiceGRPCAddr := getString("NOTIFICATION_SERVICE_GRPC_ADDR")

	rateLimitRPS := getUInt("RATE_LIMIT_RPS")
	rateLimitBucketSize := getUInt("RATE_LIMIT_BUCKET_SIZE")
	rateLimitIdleTTL := getDuration("RATE_LIMIT_IDLE_TTL")

	res, err := strconv.Atoi(os.Getenv("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS"))
	if err != nil || res == 0 {
		logger.Error("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS env missing or not a number")
		os.Exit(1)
	}
	gracefulShutdownTimeout := time.Second * time.Duration(res)

	return &Config{
		GrpcServerPort:              grpcPort,
		RestServerPort:              restPort,
		GracefulShutdownTimeout:     gracefulShutdownTimeout,
		UserServiceGRPCAddr:         userServiceGRPCAddr,
		AuthServiceGRPCAddr:         authServiceGRPCAddr,
		JwtPublicKeyPath:            jwtPublicKeyPath,
		FileServiceGRPCAddr:         fileServiceGRPCAddr,
		NotificaitonServiceGRPCAddr: notifServiceGRPCAddr,

		RateLimitRPS:        uint(rateLimitRPS),
		RateLimitBucketSize: uint(rateLimitBucketSize),
		RateLimitIdleTTL:    rateLimitIdleTTL,
	}
}
