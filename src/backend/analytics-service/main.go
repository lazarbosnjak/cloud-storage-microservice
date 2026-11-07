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

	pb "proto/analytics"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedAnalyticsServer
}

func (s *server) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingReply, error) {
	time.Sleep(time.Second * 5)
	return &pb.PingReply{Message: "Hello from analytics"}, nil
}

func main() {

	logger := log.New(os.Stdout, "[analytics server]", log.Ldate|log.Ltime)

	port := "8080"
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterAnalyticsServer(grpcServer, &server{})
	reflection.Register(grpcServer)

	sigChan := make(chan os.Signal, 1)

	go func() {
		logger.Printf("server starting on port %s", port)
		err := grpcServer.Serve(lis)
		if err != nil {
			logger.Printf("grpc could not serve on port %s", port)
		}
	}()

	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	<-sigChan

	shutdownDone := make(chan struct{})

	go func() {
		logger.Printf("Gracefully shutting down...")
		grpcServer.GracefulStop()
		shutdownDone <- struct{}{}
	}()

	select {
	case <-time.After(time.Second * 3):
		logger.Printf("Timeout has passed. Forcefully shutting down server...")
		grpcServer.Stop()
	case <-shutdownDone:
	}

	logger.Printf("Server has shut down.")

}
