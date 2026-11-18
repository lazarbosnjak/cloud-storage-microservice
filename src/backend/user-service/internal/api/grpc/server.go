package grpc

import (
	"context"
	pb "proto/user"

	"user-service/internal/application/user/register"
)

type Server struct {
	pb.UnimplementedUserServer
	CreateUserHandler CreateUserHandler
}

type CreateUserHandler interface {
	Register(ctx context.Context, query register.Query) error
}

type Handlers struct{}

func NewServer(createUserHandler CreateUserHandler) *Server {
	return &Server{
		CreateUserHandler: createUserHandler,
	}
}
