package user

import (
	"context"
	"log/slog"
	userPb "proto/user"
)

type GRPCClient struct {
	client userPb.UserClient
	logger *slog.Logger
}

func NewGRPCClient(client userPb.UserClient, logger *slog.Logger) *GRPCClient {
	return &GRPCClient{
		client: client,
		logger: logger,
	}
}

func (c *GRPCClient) Ping(
	ctx context.Context,
) (string, error) {
	reply, err := c.client.Ping(ctx, &userPb.PingRequest{})
	if err != nil {
		c.logger.Error(err.Error())
		return "", err
	}

	return reply.Message, nil
}
