package grpc

import (
	"context"
	pb "proto/api-gateway"

	userping "api-gateway/internal/application/user/ping"
)

type Server struct {
	pb.UnimplementedApiGatewayServer
	Handlers *Handlers
}

type UserPingHandler interface {
	Handle(ctx context.Context, _ userping.Query) (userping.Result, error)
}

type Handlers struct {
	UserPingHandler UserPingHandler
}

func NewServer(handlers *Handlers) *Server {
	return &Server{
		Handlers: handlers,
	}
}
