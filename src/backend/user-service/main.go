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

	pb "proto/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedUserServer
}

func (s *server) Ping(_ context.Context, in *pb.PingRequest) (*pb.PingReply, error) {
	time.Sleep(time.Second * 5)
	return &pb.PingReply{Message: "Pong " + in.FromName}, nil
}

func main() {
	logger := log.New(os.Stdout, "[user-service] ", log.Ldate|log.Ltime)

	// FIXME: port should come from an env
	port := "8080"
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		logger.Fatalf("error during tcp startup: %s\n", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterUserServer(grpcServer, &server{})
	reflection.Register(grpcServer)
	go func() {
		logger.Println("Server has started!")
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatalf("error during tcp startup: %s\n", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Println("Gracefully shutting down...")
	grpcServer.GracefulStop()
	logger.Println("Server has shut down.")
}
