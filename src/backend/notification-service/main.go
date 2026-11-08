package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	pb "proto/notification"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	grpcServerPort          = os.Getenv("GRPC_PORT")
	gracefulShutdownTimeout = os.Getenv("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS")
)

type server struct {
	pb.UnimplementedNotificationServer
}

func (s *server) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingReply, error) {
	select {
	case <-time.After(5 * time.Second):
		return &pb.PingReply{
			Message: "Pong from notification-service: " + in.FromName,
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func main() {
	logger := log.New(os.Stdout, "[notification-service] ", log.Ldate|log.Ltime)

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", grpcServerPort))
	if err != nil {
		logger.Fatalf("Failed to start listener: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterNotificationServer(grpcServer, &server{})
	reflection.Register(grpcServer)

	go func() {
		logger.Printf("Notification gRPC server running on port %s", grpcServerPort)

		if err := grpcServer.Serve(listener); err != nil {
			logger.Printf("gRPC server stopped: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	<-stop
	logger.Println("Shutdown signal received")

	shutdownDone := make(chan struct{})

	go func() {
		grpcServer.GracefulStop()
		close(shutdownDone)
	}()

	res, err := strconv.Atoi(gracefulShutdownTimeout)
	if err != nil {
		logger.Printf("can't convert gracefulShutdownTimeout to int: %v", err)
	}
	gracefulShutdownTimeout := time.Second * time.Duration(res)

	select {
	case <-shutdownDone:
		logger.Println("Notification service gracefully stopped")

	case <-time.After(gracefulShutdownTimeout):
		logger.Println("Graceful shutdown timeout reached, forcing stop")
		grpcServer.Stop()
		<-shutdownDone
		logger.Println("Notification service forcefully stopped")
	}
}
