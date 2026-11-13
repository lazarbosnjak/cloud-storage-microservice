package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "proto/auth"

	"auth-service/internal/config"
	"auth-service/internal/db"
	"auth-service/internal/handler"
	"auth-service/internal/repository"
	"auth-service/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("auth-service failed", "error", err)
		os.Exit(1)
	}

	slog.Info("service stopped")

}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	pool, err := db.CreateNewPooling(ctx, cfg.Postgres)
	if err != nil {
		return err
	}
	defer pool.Close()

	repo := repository.NewCredentialRepository(pool)
	svc := service.NewAuthService(repo)
	authHandler := handler.NewAuthHandler(svc)

	lis, err := net.Listen("tcp", ":"+cfg.Service.GRPCPort)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServer(grpcServer, authHandler)
	reflection.Register(grpcServer)

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("service started", "port", cfg.Service.GRPCPort)
		serveErr <- grpcServer.Serve(lis)
	}()

	select {
	case err := <-serveErr:
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
