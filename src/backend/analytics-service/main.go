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

	pb "proto/analytics"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	grpcServerPort          = os.Getenv("GRPC_PORT")
	gracefulShutdownTimeout = os.Getenv("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS")
)

type server struct {
	pb.UnimplementedAnalyticsServer
}

func (s *server) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingReply, error) {
	time.Sleep(time.Second * 5)
	return &pb.PingReply{Message: "hello from analytics-service"}, nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", grpcServerPort))
	if err != nil {
		logger.Error(fmt.Sprintf("failed to listen: %v", err))
	}

	grpcServer := grpc.NewServer()

	pb.RegisterAnalyticsServer(grpcServer, &server{})
	reflection.Register(grpcServer)

	go func() {
		logger.Info(fmt.Sprintf("serice starting on port: %s", grpcServerPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error(fmt.Sprintf("grpc could not serve on port: %s", grpcServerPort))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	<-sigChan

	shutdownDone := make(chan struct{})

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
		logger.Warn("timeout has passed. Forcefully shutting down...")
		grpcServer.Stop()
	case <-shutdownDone:
	}

	logger.Info("service is stopped.")
}
