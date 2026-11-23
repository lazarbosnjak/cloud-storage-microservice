package grpc

import (
	pb "proto/api-gateway"

	authlogin "api-gateway/internal/application/auth/login"
	authregister "api-gateway/internal/application/auth/register"
	userping "api-gateway/internal/application/user/ping"
)

type Server struct {
	pb.UnimplementedApiGatewayServer
	Handlers *Handlers
}

type Handlers struct {
	UserPingHandler     *userping.Handler
	AuthRegisterHandler *authregister.Handler
	AuthLoginHandler    *authlogin.Handler
}

func NewServer(handlers *Handlers) *Server {
	return &Server{
		Handlers: handlers,
	}
}
