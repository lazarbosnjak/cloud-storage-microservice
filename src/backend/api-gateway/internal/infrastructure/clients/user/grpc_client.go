package user

import (
	"context"
	"log/slog"

	appuser "api-gateway/internal/application/user"
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

func (c *GRPCClient) Ping(ctx context.Context) (string, error) {
	reply, err := c.client.Ping(ctx, &userPb.PingRequest{})
	if err != nil {
		c.logger.Error(err.Error())
		return "", err
	}

	return reply.Message, nil
}

func (c *GRPCClient) GetStorage(ctx context.Context, userID string) (*appuser.Storage, error) {
	reply, err := c.client.GetStorage(ctx, &userPb.GetStorageRequest{
		UserId: userID,
	})
	if err != nil {
		c.logger.Error(err.Error())
		return nil, err
	}

	return &appuser.Storage{
		Quota:     reply.GetQuota(),
		Used:      reply.GetUsed(),
		Available: reply.GetAvailable(),
	}, nil
}
