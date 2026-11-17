package startup

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"user-service/internal/startup/config"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "proto/user"
	usergrpc "user-service/internal/api/grpc"
)

type Server struct {
	config *config.Config
	logger *slog.Logger
}

func NewServer(config *config.Config, logger *slog.Logger) *Server {
	return &Server{config: config, logger: logger}
}

func (s *Server) Start() {
	if err := s.run(); err != nil {
		s.logger.Error("server exited with error", "err", err)
		os.Exit(1)
	}
	s.logger.Info("service stopped")
}

func (s *Server) run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	grpcAPIServer := s.newAPIServer()

	grpcServer := grpc.NewServer()
	pb.RegisterUserServer(grpcServer, grpcAPIServer)
	reflection.Register(grpcServer)

	grpcAddr := ":" + s.config.GrpcPort
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", grpcAddr, err)
	}

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		s.logger.Info("gRPC server started", "port", s.config.GrpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			return fmt.Errorf("grpc serve: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		<-gctx.Done()
		s.shutdown(grpcServer)
		return nil
	})

	return g.Wait()
}

func (s *Server) shutdown(grpcServer *grpc.Server) {
	s.logger.Info("gracefully shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), s.config.GracefulShutdownTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-ctx.Done():
		s.logger.Warn("graceful shutdown timeout exceeded; forcing stop...")
		grpcServer.Stop()
	}
}

func (s *Server) newAPIServer() pb.UserServer {
	return usergrpc.NewServer()
}
