package grpc

import (
	"context"
	"fmt"
	pb "proto/user"

	createuser "user-service/internal/application/user/create"
)

func (s *Server) CreateUser(ctx context.Context, in *pb.CreateUserRequest) (*pb.CreateUserReply, error) {
	query := &createuser.Query{
		ID:        in.Id,
		Email:     in.Email,
		FirstName: in.FirstName,
		LastName:  in.LastName,
	}

	reply := &pb.CreateUserReply{}

	err := s.CreateUserHandler.Register(ctx, *query)
	if err != nil {
		return reply, fmt.Errorf("gRPC CreateUser error: %w", err)
	}

	return reply, nil
}
