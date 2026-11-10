package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	pb "proto/api-gateway"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

var (
	grpcServerPort          = os.Getenv("GRPC_PORT")
	restServerPort          = os.Getenv("REST_PORT")
	gracefulShutdownTimeout = os.Getenv("GRACEFUL_SHUTDOWN_TIMEOUT_SECONDS")
)

type server struct {
	pb.UnimplementedApiGatewayServer
}

func (s *server) Ping(_ context.Context, in *pb.PingRequest) (*pb.PingReply, error) {
	time.Sleep(time.Second * 5)
	return &pb.PingReply{Message: "hello from api-gateway"}, nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", grpcServerPort))
	if err != nil {
		logger.Error(fmt.Sprintf("failed to listen: %v", err))
	}

	grpcServer := grpc.NewServer()

	pb.RegisterApiGatewayServer(grpcServer, &server{})
	reflection.Register(grpcServer)
	go func() {
		logger.Info(fmt.Sprintf("gRPC server started on port: %s", grpcServerPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error(fmt.Sprintf("grpc could not serve on port: %s", grpcServerPort))
		}
	}()

	conn, err := grpc.NewClient(
		"0.0.0.0:"+grpcServerPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("error creating gRPC client", "err", err)
	}

	gwMux := runtime.NewServeMux()
	err = pb.RegisterApiGatewayHandler(ctx, gwMux, conn)
	if err != nil {
		logger.Error("failed to register gateway", "err", err)
	}

	gwServer := &http.Server{
		Addr:    ":" + restServerPort,
		Handler: gwMux,
	}

	go func() {
		logger.Info(fmt.Sprintf("REST server started on port: %s", restServerPort))
		err := gwServer.ListenAndServe()
		if err != nil {
			logger.Error(err.Error())
		}
	}()

	// Gracecful shutdown

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
