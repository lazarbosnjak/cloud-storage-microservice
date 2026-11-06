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

	pb "proto/sharing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedSharingServer
}

func (s *server) Ping(_ context.Context, _ *pb.PingRequest) (*pb.PingReply, error) {
	time.Sleep(time.Second * 5)
	return &pb.PingReply{Message: "Pong from sharing-service"}, nil
}

func main() {
	// TODO: replace port with an env
	// port := os.Getenv("SHARING_SERVICE_PORT")

	logger := log.New(os.Stdout, "[sharing-service] ", log.Ldate|log.Ltime)

	port := "8080"
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		logger.Fatalf("could not listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSharingServer(grpcServer, &server{})
	reflection.Register(grpcServer)

	go func() {
		logger.Printf("Server started on port: %s\n", port)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatalf("grpcServer error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	logger.Println("Gracefully shutting down...")
	grpcServer.GracefulStop()
	logger.Println("Server has shut down.")
}
