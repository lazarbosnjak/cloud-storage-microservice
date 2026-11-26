package notification

import (
	"context"
	pb "proto/notification"
)

type GRPCClient struct {
	client pb.NotificationClient
}

func NewGRPCClient(client pb.NotificationClient) *GRPCClient {
	return &GRPCClient{
		client: client,
	}
}

func (c *GRPCClient) GetAllByUser(
	ctx context.Context,
	in *pb.GetAllByUserRequest,
) (*pb.GetAllByUserResponse, error) {
	return c.client.GetAllByUser(ctx, in)
}
