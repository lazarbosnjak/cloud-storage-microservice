package grpc

import (
	"context"
	pb "proto/user"

	createuser "user-service/internal/application/user/create"
)

type Server struct {
	pb.UnimplementedUserServer
	CreateUserHandler CreateUserHandler
}

type CreateUserHandler interface {
	Register(ctx context.Context, query createuser.Query) error
}

type Handlers struct{}

func NewServer(createUserHandler CreateUserHandler) *Server {
	return &Server{
		CreateUserHandler: createUserHandler,
	}
}
