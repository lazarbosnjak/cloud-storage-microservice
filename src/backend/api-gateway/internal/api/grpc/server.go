package grpc

import (
	"context"
	pb "proto/api-gateway"

	authregister "api-gateway/internal/application/auth/register"
	userping "api-gateway/internal/application/user/ping"
)

type Server struct {
	pb.UnimplementedApiGatewayServer
	Handlers *Handlers
}

type UserPingHandler interface {
	Handle(ctx context.Context, _ userping.Query) (userping.Result, error)
}

type AuthRegisterHandler interface {
	Handle(ctx context.Context, command authregister.Command) error
}

type Handlers struct {
	UserPingHandler     UserPingHandler
	AuthRegisterHandler AuthRegisterHandler
}

func NewServer(handlers *Handlers) *Server {
	return &Server{
		Handlers: handlers,
	}
}
