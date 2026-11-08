package main

import (
	"context"
	"fmt"
	"log"
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
	return &pb.PingReply{Message: "Hello from analytics"}, nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", grpcServerPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterAnalyticsServer(grpcServer, &server{})
	reflection.Register(grpcServer)

	sigChan := make(chan os.Signal, 1)

	go func() {
		logger.Info(fmt.Sprintf("server starting on port %s", grpcServerPort))
		err := grpcServer.Serve(lis)
		if err != nil {
			logger.Error(fmt.Sprintf("grpc could not serve on port %s", grpcServerPort))
		}
	}()

	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	<-sigChan

	shutdownDone := make(chan struct{})

	go func() {
		logger.Info("Gracefully shutting down...")
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
		logger.Warn("Timeout has passed. Forcefully shutting down server...")
		grpcServer.Stop()
	case <-shutdownDone:
	}

	logger.Info("Server has shut down.")
}
