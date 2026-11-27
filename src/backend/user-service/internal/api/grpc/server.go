package grpc

import (
	"context"
	pb "proto/user"

	createuser "user-service/internal/application/user/create"
	"user-service/internal/domain"
)

type Server struct {
	pb.UnimplementedUserServer
	CreateUserHandler CreateUserHandler
	StorageHandler    StorageHandler
}

type CreateUserHandler interface {
	Register(ctx context.Context, query createuser.Query) error
}

type StorageHandler interface {
	GetStorage(
		ctx context.Context,
		userID string,
	) (domain.StorageInfo, error)

	AdjustStorageUsage(
		ctx context.Context,
		userID string,
		delta int64,
	) (domain.StorageInfo, error)
}

type Handlers struct{}

func NewServer(
	createUserHandler CreateUserHandler,
	storageHandler StorageHandler,
) *Server {
	return &Server{
		CreateUserHandler: createUserHandler,
		StorageHandler:    storageHandler,
	}
}
