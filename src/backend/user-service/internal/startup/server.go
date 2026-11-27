package startup

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	createuser "user-service/internal/application/user/create"
	"user-service/internal/infrastructure/mongodb"
	"user-service/internal/startup/config"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "proto/user"
	usergrpc "user-service/internal/api/grpc"
	storageapp "user-service/internal/application/user/storage"
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

	grpcAPIServer := s.newAPIServer(ctx)

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

func (s *Server) newAPIServer(ctx context.Context) pb.UserServer {
	// Mongo client
	mongoClient, err := mongodb.NewClient(ctx, s.config)
	if err != nil {
		s.logger.Error("error creating newAPIServer", "err", err.Error())
		os.Exit(1)
	}

	// Collections
	userRepo := mongodb.NewUserRepository(
		mongoClient.Database(s.config.Mongo.Database),
		s.logger)

	// Handlers
	createUserHandler := createuser.NewHandler(userRepo, s.config.DefaultStorageQuota)
	storageHandler := storageapp.NewHandler(userRepo)

	return usergrpc.NewServer(createUserHandler, storageHandler)
}
