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
	time.Sleep(time.Second * 5)
	return &pb.PingReply{Message: "hello from notification-service"}, nil
	/*select {
	case <-time.After(5 * time.Second):
		return &pb.PingReply{
			Message: "hello from notification-service: " + in.FromName,
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}*/
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", grpcServerPort))
	if err != nil {
		logger.Error(fmt.Sprintf("failed to listen: %v", err))
	}

	grpcServer := grpc.NewServer()

	pb.RegisterNotificationServer(grpcServer, &server{})
	reflection.Register(grpcServer)

	go func() {
		logger.Info(fmt.Sprintf("service running on port: %s", grpcServerPort))
		if err := grpcServer.Serve(listener); err != nil {
			logger.Error(fmt.Sprintf("grpc could not serve on port: %v", err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	<-stop

	shutdownDone := make(chan struct{})

	go func() {
		logger.Info("gracefully shutting down...")
		grpcServer.GracefulStop()
		close(shutdownDone)
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
