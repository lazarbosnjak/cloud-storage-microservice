package auth

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

	pb "proto/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	grpcServerPort          = os.Getenv("GRPC_PORT")
	gracefulShutdownTimeout = os.Getenv("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS")
)

type server struct {
	pb.UnimplementedAuthServer
}

func (s *server) Ping(_ context.Context, in *pb.PingRequest) (*pb.PingReply, error) {
	time.Sleep(time.Second * 5)
	return &pb.PingReply{Message: "Pong from auth-service"}, nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", grpcServerPort))
	if err != nil {
		logger.Error(fmt.Sprintf("failed to listen: %v", err))
	}

	grpcServer := grpc.NewServer()

	pb.RegisterAuthServer(grpcServer, &server{})
	reflection.Register(grpcServer)
	go func() {
		logger.Info(fmt.Sprintf("service started on port: %s", grpcServerPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error(fmt.Sprintf("grpc could not serve on port: %s", grpcServerPort))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	shutdownDone := make(chan struct{}, 1)

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
		logger.Warn("graceful shutdown timeout exceeded; forcefully shutting down...")
		grpcServer.Stop()
	case <-shutdownDone:
	}

	logger.Info("service stopped.")
}
