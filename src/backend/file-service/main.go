package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "proto/file"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedFileServer
}

func (s *server) Ping(_ context.Context, _ *pb.PingRequest) (*pb.PingReply, error) {
	time.Sleep(time.Second * 5)
	return &pb.PingReply{Message: "Pong from file-service"}, nil
}

func main() {
	logger := log.New(os.Stdout, "[file-service] ", log.Ldate|log.Ltime)

	// TODO: replace port with env
	// port := os.Getenv("FILE_SERVICE_PORT")

	port := "8080"
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		logger.Fatalf("error during tcp startup: %s\n", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterFileServer(grpcServer, &server{})
	reflection.Register(grpcServer)
	go func() {
		logger.Printf("Server has started on port: %s\n", port)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatalf("grpc server error: %s\n", err)
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

	// TODO: Change to env
	const GracefulShutdownTimeout = time.Second * 3

	select {
	case <-shutdownDone:
	case <-time.After(GracefulShutdownTimeout):
		logger.Println("Graceful shutdown timeout reached. Shutting down forcefully...")
		grpcServer.Stop()
	}
	logger.Println("Server has shut down.")
}
