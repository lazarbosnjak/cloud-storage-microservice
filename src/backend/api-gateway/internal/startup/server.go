package startup

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	gatewaygrpc "api-gateway/internal/api/grpc"
	"api-gateway/internal/application/auth/login"
	"api-gateway/internal/application/auth/register"
	"api-gateway/internal/application/user/ping"
	"api-gateway/internal/infrastructure/clients/auth"
	"api-gateway/internal/infrastructure/clients/user"
	"api-gateway/internal/startup/config"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	pb "proto/api-gateway"
	authPb "proto/auth"
	userPb "proto/user"
)

type Server struct {
	config *config.Config
	logger *slog.Logger
}

func NewServer(config *config.Config, logger *slog.Logger) *Server {
	return &Server{config: config, logger: logger}
}

func (s *Server) Start() {
	if err := s.run(); err != nil {
		s.logger.Error("server exited with error", "err", err)
		os.Exit(1)
	}
	s.logger.Info("service stopped")
}

func (s *Server) run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Dependencies
	userConn, err := grpc.NewClient(
		s.config.UserServiceGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("create user gRPC connection: %w", err)
	}
	defer func() {
		if err := userConn.Close(); err != nil {
			s.logger.Error("close connection ")
		}
	}()

	authConn, err := grpc.NewClient(
		s.config.AuthServiceGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("create auth gRPC connection: %w", err)
	}
	defer func() {
		if err := authConn.Close(); err != nil {
			s.logger.Error("close connection ")
		}
	}()

	grpcAPIServer := s.newAPIServer(userConn, authConn)

	// gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterApiGatewayServer(grpcServer, grpcAPIServer)
	reflection.Register(grpcServer)

	grpcAddr := ":" + s.config.GrpcServerPort
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", grpcAddr, err)
	}

	// REST gateway
	restConn, err := grpc.NewClient(
		"localhost:"+s.config.GrpcServerPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("create gateway client client: %w", err)
	}
	defer func() {
		if err := restConn.Close(); err != nil {
			s.logger.Error("close connection ")
		}
	}()

	cookieSetter := runtime.WithOutgoingHeaderMatcher(func(key string) (string, bool) {
		if strings.EqualFold(key, "set-cookie") {
			return "Set-Cookie", true
		}
		return runtime.DefaultHeaderMatcher(key)
	})

	gwMux := runtime.NewServeMux(
		cookieSetter,
	)
	if err := pb.RegisterApiGatewayHandler(ctx, gwMux, restConn); err != nil {
		return fmt.Errorf("register gateway handler: %w", err)
	}

	gwServer := &http.Server{
		Addr:    ":" + s.config.RestServerPort,
		Handler: gwMux,
	}

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		s.logger.Info("gRPC server started", "port", s.config.GrpcServerPort)
		if err := grpcServer.Serve(lis); err != nil {
			return fmt.Errorf("grpc serve: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		s.logger.Info("REST server started", "port", s.config.RestServerPort)
		if err := gwServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("rest server error: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		<-gctx.Done()
		s.shutdown(gwServer, grpcServer)
		return nil
	})

	return g.Wait()
}

func (s *Server) shutdown(gwServer *http.Server, grpcServer *grpc.Server) {
	s.logger.Info("gracefully shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), s.config.GracefulShutdownTimeout)
	defer cancel()

	if err := gwServer.Shutdown(ctx); err != nil {
		s.logger.Error("failed to shut down REST server", "err", err)
	}

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-ctx.Done():
		s.logger.Warn("graceful shutdown timeout exceeded; forcing stop...")
		grpcServer.Stop()
	}
}

func (s *Server) newAPIServer(
	userConn *grpc.ClientConn,
	authConn *grpc.ClientConn,
) pb.ApiGatewayServer {
	userClient := user.NewGRPCClient(userPb.NewUserClient(userConn), s.logger)
	userPingHandler := ping.NewHandler(userClient)

	authClient := auth.NewGRPCClient(authPb.NewAuthClient(authConn), s.logger)
	authRegisterHandler := register.NewHandler(authClient)
	authLoginHandler := login.NewHandler(authClient)

	return gatewaygrpc.NewServer(&gatewaygrpc.Handlers{
		UserPingHandler:     userPingHandler,
		AuthRegisterHandler: authRegisterHandler,
		AuthLoginHandler:    authLoginHandler,
	})
}
