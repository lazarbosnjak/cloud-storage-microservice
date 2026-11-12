package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	pb "proto/file"

	"file-service/database"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	grpcServerPort          = os.Getenv("GRPC_PORT")
	gracefulShutdownTimeout = os.Getenv("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS")
)

type server struct {
	pb.UnimplementedFileServer
}

func (s *server) Ping(_ context.Context, _ *pb.PingRequest) (*pb.PingReply, error) {
	time.Sleep(time.Second * 5)
	return &pb.PingReply{Message: "hello from file-service"}, nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	//veza sa bazom
	mongoClient, err := database.ConnectMongoDB(context.Background())
	if err != nil {
		logger.Error("failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}

	logger.Info("successfully connected to MongoDB")

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", grpcServerPort))
	if err != nil {
		logger.Error(fmt.Sprintf("failed to listen: %v", err))
	}

	grpcServer := grpc.NewServer()

	pb.RegisterFileServer(grpcServer, &server{})
	reflection.Register(grpcServer)
	go func() {
		logger.Info(fmt.Sprintf("service started on port: %s", grpcServerPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error(fmt.Sprintf("grpc could not serve on port: %s", grpcServerPort))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	shutdownDone := make(chan struct{}, 1)

	go func() {
		logger.Info("gracefully shutting down...")
		grpcServer.GracefulStop()
		shutdownDone <- struct{}{}
	}()

	res, err := strconv.Atoi(gracefulShutdownTimeout)
	if err != nil {
		logger.Error(fmt.Sprintf("can't convert gracefulShutdownTimeout to int: %v", err))
	}
	gracefulShutdownTimeout := time.Second * time.Duration(res)

	select {
	case <-time.After(gracefulShutdownTimeout):
		logger.Warn("graceful shutdown timeout exceeded; forcefully shutting down...")
		grpcServer.Stop()
	case <-shutdownDone:
	}

	//zatvaranje konekcije pri gasenju servisa
	mongoCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := mongoClient.Disconnect(mongoCtx); err != nil {
		logger.Error("failed to disconnect from MongoDB", "error", err)
	} else {
		logger.Info("MongoDB connection closed successfully")
	}

	logger.Info("service stopped.")

}
