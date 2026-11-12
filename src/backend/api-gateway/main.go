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

	"api-gateway/internal/application/user/ping"
	"api-gateway/internal/infrastructure/clients/user"
	grpcapi "api-gateway/internal/transport/grpc"

	pb "proto/api-gateway"
	userPb "proto/user"

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

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", grpcServerPort))
	if err != nil {
		logger.Error(fmt.Sprintf("failed to listen: %v", err))
	}

	grpcServer := grpc.NewServer()

	// TODO: put user-service addr in env
	userConn, err := grpc.NewClient(
		"user-service:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error(err.Error())
	}
	userRPCClient := userPb.NewUserClient(userConn)
	userClient := user.NewGRPCClient(userRPCClient, logger)
	pingUserHandler := ping.NewHandler(userClient)

	server := grpcapi.NewServer(pingUserHandler)

	pb.RegisterApiGatewayServer(grpcServer, server)
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
