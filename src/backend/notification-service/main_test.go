package main

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestGracefulShutdown(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	server := grpc.NewServer()
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)

	serverStopped := make(chan struct{})

	go func() {
		defer close(serverStopped)
		_ = server.Serve(listener)
	}()

	conn, err := grpc.NewClient(
		listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := healthpb.NewHealthClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = client.Check(ctx, &healthpb.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("gRPC request failed: %v", err)
	}

	shutdownDone := make(chan struct{})

	go func() {
		server.GracefulStop()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		t.Log("Graceful shutdown completed successfully")
	case <-time.After(5 * time.Second):
		server.Stop()
		t.Fatal("Graceful shutdown timed out")
	}

	select {
	case <-serverStopped:
		t.Log("gRPC server stopped successfully")
	case <-time.After(5 * time.Second):
		t.Fatal("gRPC Serve did not stop")
	}
}

func TestGracefulShutdownWaitsForActiveRequest(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	server := grpc.NewServer()
	requestStarted := make(chan struct{})
	finishRequest := make(chan struct{})

	server.RegisterService(&grpc.ServiceDesc{
		ServiceName: "test.ShutdownService",
		HandlerType: (*shutdownTestService)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "Wait",
				Handler: func(
					srv interface{},
					ctx context.Context,
					dec func(interface{}) error,
					interceptor grpc.UnaryServerInterceptor,
				) (interface{}, error) {
					req := &emptypb.Empty{}
					if err := dec(req); err != nil {
						return nil, err
					}

					handler := func(
						ctx context.Context,
						req interface{},
					) (interface{}, error) {
						close(requestStarted)

						select {
						case <-finishRequest:
							return &emptypb.Empty{}, nil
						case <-ctx.Done():
							return nil, ctx.Err()
						}
					}

					if interceptor == nil {
						return handler(ctx, req)
					}

					return interceptor(ctx, req, &grpc.UnaryServerInfo{
						Server:     srv,
						FullMethod: "/test.ShutdownService/Wait",
					}, handler)
				},
			},
		},
	}, &shutdownTestServiceImpl{})

	go func() {
		_ = server.Serve(listener)
	}()

	conn, err := grpc.NewClient(
		listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		server.Stop()
		t.Fatal(err)
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	requestDone := make(chan error, 1)

	go func() {
		err := conn.Invoke(
			ctx,
			"/test.ShutdownService/Wait",
			&emptypb.Empty{},
			&emptypb.Empty{},
		)
		requestDone <- err
	}()

	select {
	case <-requestStarted:
		t.Log("RPC request started")
	case <-ctx.Done():
		server.Stop()
		t.Fatal("RPC request did not start")
	}

	shutdownDone := make(chan struct{})

	go func() {
		server.GracefulStop()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		t.Fatal("Server stopped before RPC finished")
	case <-time.After(200 * time.Millisecond):
		t.Log("GracefulStop is waiting for active RPC")
	}

	close(finishRequest)

	select {
	case err := <-requestDone:
		if err != nil {
			t.Fatalf("RPC failed: %v", err)
		}
		t.Log("Active RPC completed successfully")
	case <-ctx.Done():
		server.Stop()
		t.Fatal("RPC did not finish")
	}

	select {
	case <-shutdownDone:
		t.Log("Graceful shutdown completed after RPC")
	case <-ctx.Done():
		server.Stop()
		t.Fatal("Graceful shutdown timed out")
	}
}

type shutdownTestService interface{}

type shutdownTestServiceImpl struct{}
