package startup

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	notificationgrpc "notification-service/internal/api/grpc"
	getbyuser "notification-service/internal/application/notification/get_by_user"
	"notification-service/internal/infrastructure/db/cassandra"
	"notification-service/internal/startup/config"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "proto/notification"
)

type Server struct {
	cfg    *config.Config
	logger *slog.Logger
}

func NewServer(cfg *config.Config, logger *slog.Logger) *Server {
	return &Server{
		cfg:    cfg,
		logger: logger,
	}
}

func (s *Server) Start() {
	if err := s.run(); err != nil {
		s.logger.Error("error initiating the server", "err", err)
		os.Exit(1)
	}
	s.logger.Info("service stopped running")
}

func (s *Server) run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	grpcServer := grpc.NewServer()

	notifGRPCServer, err := s.newGRPCServer()
	if err != nil {
		return err
	}

	pb.RegisterNotificationServer(grpcServer, notifGRPCServer)
	reflection.Register(grpcServer)

	grpcAddr := ":" + s.cfg.GrpcPort
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", grpcAddr, err)
	}

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		s.logger.Info("gRPC server started", "port", s.cfg.GrpcPort)
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

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.GracefulShutdownTimeout)
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

func (s *Server) newGRPCServer() (pb.NotificationServer, error) {
	notifRepo, err := cassandra.NewNotificationsRepository(s.cfg)
	if err != nil {
		return nil, err
	}

	notifGetByUserHandler := getbyuser.NewHandler(notifRepo)

	return notificationgrpc.NewServer(
		notifGetByUserHandler,
	), nil
}
