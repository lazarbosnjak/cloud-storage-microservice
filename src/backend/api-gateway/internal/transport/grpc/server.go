package grpc

import (
	pb "proto/api-gateway"

	"api-gateway/internal/application/user/ping"
)

type server struct {
	pb.UnimplementedApiGatewayServer
	pingUserHandler *ping.Handler
}

func NewServer(pingUserHandler *ping.Handler) *server {
	return &server{
		pingUserHandler: pingUserHandler,
	}
}
