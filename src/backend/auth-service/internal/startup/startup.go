package startup

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth-service/internal/db"
	"auth-service/internal/handler"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"auth-service/internal/startup/config"

	pb "proto/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type Service struct {
	cfg *config.Config
}

func Start() {
	if err := run(); err != nil {
		slog.Error("service exited with error", "error", err)
		os.Exit(1)
	}
	slog.Info("service stopped")
}

// FIXME:
// Make this a method of *Service so you
// get access to configurations through it
func run() error {
	// Setting up KILL signal
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Loading configurations
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		return err
	}

	// Initializing connection pool
	pool, err := db.CreateNewPooling(ctx, cfg.Postgres)
	if err != nil {
		slog.Error("create pooling", "error", err)
		return err
	}
	defer pool.Close()

	// If database is new, we will automatically create table
	if err := db.RunMigrations(cfg.Postgres.GetDataSourceName()); err != nil {
		slog.Error("running migrations", "error", err)
		return err
	}

	// Dependencies
	userConn, err := grpc.NewClient(
		cfg.Service.UserServiceGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("create user gRPC connection: %w", err)
	}
	defer func() {
		if err := userConn.Close(); err != nil {
			slog.Error("close connection error")
		}
	}()

	// Connecting everything
	repo := repository.NewCredentialRepository(pool)
	userGRPCClient := handler.NewGRPCUserClient(userConn)
	svc := service.NewAuthService(repo, cfg, userGRPCClient)
	authHandler := handler.NewAuthHandler(svc)

	// Adding a default admin account. Used mainly in testing purposes
	if cfg.Service.AdminEmail != "" {
		if err := svc.EnsureAdmin(ctx, cfg.Service.AdminEmail, cfg.Service.AdminPassword); err != nil {
			slog.Error("bootstrap admin", "error", err)
			return fmt.Errorf("error while bootstraping admin: %w", err)
		}
	}

	// Basic listener for our grpc port
	lis, err := net.Listen("tcp", ":"+cfg.Service.GRPCPort)
	if err != nil {
		slog.Error("grcp listen", "error", err)
		return err
	}

	// Setting up grpc server and allowing direct access to external network via grpc
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServer(grpcServer, authHandler)
	reflection.Register(grpcServer)

	// Rest of the code is mainly starting services correctly and gracefully shutting them down if
	serveErr := make(chan error, 1)
	go func() {
		slog.Info("service started", "port", cfg.Service.GRPCPort)
		serveErr <- grpcServer.Serve(lis)
	}()

	select {
	case err := <-serveErr:
		slog.Error("starting grcp", "error", err)
		return err
	case <-ctx.Done():
		slog.Info("gracefully shutting down...")
	}

	shutdownDone := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
	case <-time.After(time.Duration(cfg.Service.GracefulShutdownTimeout)):
		slog.Warn("graceful shutdown timeout exceeded; forcing stop")
		grpcServer.Stop()
	}

	return nil
}
