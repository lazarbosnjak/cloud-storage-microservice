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

	pb "proto/sharing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	grpcServerPort          = os.Getenv("GRPC_PORT")
	gracefulShutdownTimeout = os.Getenv("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS")
)

type server struct {
	pb.UnimplementedSharingServer
}

func (s *server) Ping(_ context.Context, _ *pb.PingRequest) (*pb.PingReply, error) {
	time.Sleep(time.Second * 5)
	return &pb.PingReply{Message: "Pong from sharing-service"}, nil
}

func main() {
	logger := log.New(os.Stdout, "[sharing-service] ", log.Ldate|log.Ltime)

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", grpcServerPort))
	if err != nil {
		logger.Fatalf("could not listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSharingServer(grpcServer, &server{})
	reflection.Register(grpcServer)

	go func() {
		logger.Printf("Server started on port: %s\n", grpcServerPort)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatalf("grpcServer error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	shutdownDone := make(chan struct{}, 1)

	go func() {
		logger.Println("Gracefully shutting down...")
		grpcServer.GracefulStop()
		shutdownDone <- struct{}{}
	}()

	res, err := strconv.Atoi(gracefulShutdownTimeout)
	if err != nil {
		logger.Printf("can't convert gracefulShutdownTimeout to int: %v", err)
	}
	gracefulShutdownTimeout := time.Second * time.Duration(res)

	select {
	case <-time.After(gracefulShutdownTimeout):
		logger.Println("Graceful shutdown timeout reached. Shutting down forcefully...")
		grpcServer.Stop()
	case <-shutdownDone:
	}
	logger.Println("Server has shut down.")
}
