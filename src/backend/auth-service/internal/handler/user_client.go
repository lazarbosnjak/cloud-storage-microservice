package handler

import (
	"context"
	userpb "proto/user"

	"auth-service/internal/service"

	"google.golang.org/grpc"
)

type UserClient struct {
	client userpb.UserClient
}

func NewGRPCUserClient(userConn *grpc.ClientConn) *UserClient {
	return &UserClient{
		client: userpb.NewUserClient(userConn),
	}
}

func (c *UserClient) CreateUser(ctx context.Context, user *service.CreateUserDTO) error {
	req := &userpb.CreateUserRequest{
		Id:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	_, err := c.client.CreateUser(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
