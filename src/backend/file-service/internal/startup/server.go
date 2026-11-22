package startup

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os/signal"
	"syscall"
	"time"

	"file-service/internal/application/file"
	"file-service/internal/application/folder"
	"file-service/internal/application/ping"
	"file-service/internal/infrastructure/database"
	infrarepo "file-service/internal/infrastructure/repository"
	"file-service/internal/startup/config"
	transportgrpc "file-service/internal/transport/grpc"

	pb "proto/file"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	config *config.Config
	logger *slog.Logger
}

func NewServer(cfg *config.Config, logger *slog.Logger) *Server {
	return &Server{
		config: cfg,
		logger: logger,
	}
}

func (s *Server) Start() error {
	return s.run()
}

func (s *Server) run() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	mongoClient, err := database.ConnectMongoDB(ctx, s.config)
	if err != nil {
		return fmt.Errorf("connect to MongoDB: %w", err)
	}

	s.logger.Info("successfully connected to MongoDB")

	defer func() {
		disconnectCtx, cancel := context.WithTimeout(
			context.Background(),
			s.config.MongoDisconnectTimeout,
		)
		defer cancel()

		if err := mongoClient.Disconnect(disconnectCtx); err != nil {
			s.logger.Error("failed to disconnect from MongoDB", "err", err)
		} else {
			s.logger.Info("MongoDB connection closed successfully")
		}
	}()

	listener, err := net.Listen("tcp", ":"+s.config.GRPCPort)
	if err != nil {
		return fmt.Errorf("start gRPC listener: %w", err)
	}
	defer listener.Close()

	grpcServer := grpc.NewServer()
	db := mongoClient.Database(s.config.MongoDB)

	folderRepository := infrarepo.NewFolderRepository(db)
	fileRepository := infrarepo.NewFileRepository(db)

	pingHandler := ping.NewHandler(s.config.PingDelay)
	createFolderHandler := folder.NewCreateHandler(folderRepository)

	renameFileHandler := file.NewRenameHandler(fileRepository)
	renameFolderHandler := folder.NewRenameHandler(folderRepository)

	listContentsHandler := folder.NewListContentsHandler(
		folderRepository,
		fileRepository,
	)

	getFileHandler := file.NewGetHandler(fileRepository)
	getFolderHandler := folder.NewGetHandler(folderRepository)

	pb.RegisterFileServer(
		grpcServer,
		transportgrpc.NewServer(
			pingHandler,
			createFolderHandler,
			renameFileHandler,
			renameFolderHandler,
			listContentsHandler,
			getFileHandler,
			getFolderHandler,
		),
	)

	reflection.Register(grpcServer)

	serveErrors := make(chan error, 1)

	go func() {
		s.logger.Info("gRPC server started", "port", s.config.GRPCPort)

		if err := grpcServer.Serve(listener); err != nil {
			serveErrors <- err
		} else {
			serveErrors <- nil
		}
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("shutdown signal received")

	case err := <-serveErrors:
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("gRPC server failed: %w", err)
		}
		return nil
	}

	s.shutdown(grpcServer)

	s.logger.Info("service stopped")
	return nil
}

func (s *Server) shutdown(grpcServer *grpc.Server) {
	s.logger.Info("gracefully shutting down...")

	done := make(chan struct{})

	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	timer := time.NewTimer(s.config.GracefulShutdownTimeout)
	defer timer.Stop()

	select {
	case <-done:
		s.logger.Info("gRPC server stopped gracefully")

	case <-timer.C:
		s.logger.Warn("graceful shutdown timeout exceeded; forcing stop")
		grpcServer.Stop()
		<-done
	}
}
