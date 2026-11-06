package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "proto/notification"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	port := "50051"

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Fatalf("Failed to start listener: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterNotificationServer(grpcServer, &server{})
	reflection.Register(grpcServer)

	go func() {
		logger.Printf("Notification gRPC server running on port %s", port)

		if err := grpcServer.Serve(listener); err != nil {
			logger.Printf("gRPC server stopped: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	<-stop
	logger.Println("Shutdown signal received")

	// 10 sekundi da li treba?
	shutdownDone := make(chan struct{})

	go func() {
		grpcServer.GracefulStop()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		logger.Println("Notification service gracefully stopped")

	case <-time.After(10 * time.Second):
		logger.Println("Graceful shutdown timeout reached, forcing stop")
		grpcServer.Stop()
		<-shutdownDone
		logger.Println("Notification service forcefully stopped")
	}
}
