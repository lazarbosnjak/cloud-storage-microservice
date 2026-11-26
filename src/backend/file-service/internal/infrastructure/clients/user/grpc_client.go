package user

import (
	"context"
	"fmt"
	"time"

	"file-service/internal/domain/quota"

	pb "proto/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	connection *grpc.ClientConn
	client     pb.UserClient
	timeout    time.Duration
}

var _ quota.Client = (*Client)(nil)

func NewClient(
	address string,
	timeout time.Duration,
) (*Client, error) {
	connection, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		return nil, fmt.Errorf(
			"create user service client: %w",
			err,
		)
	}

	return &Client{
		connection: connection,
		client:     pb.NewUserClient(connection),
		timeout:    timeout,
	}, nil
}

func (c *Client) Close() error {
	return c.connection.Close()
}

func (c *Client) GetStorage(
	ctx context.Context,
	userID string,
) (quota.StorageInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	reply, err := c.client.GetStorage(
		ctx,
		&pb.GetStorageRequest{
			UserId: userID,
		},
	)

	if err != nil {
		return quota.StorageInfo{}, mapError(err)
	}

	return toStorageInfo(reply), nil
}

func (c *Client) AdjustStorageUsage(
	ctx context.Context,
	userID string,
	delta int64,
) (quota.StorageInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	reply, err := c.client.AdjustStorageUsage(
		ctx,
		&pb.AdjustStorageUsageRequest{
			UserId: userID,
			Delta:  delta,
		},
	)

	if err != nil {
		return quota.StorageInfo{}, mapError(err)
	}

	return toStorageInfo(reply), nil
}

func toStorageInfo(
	reply *pb.StorageReply,
) quota.StorageInfo {
	return quota.StorageInfo{
		Quota:     reply.GetQuota(),
		Used:      reply.GetUsed(),
		Available: reply.GetAvailable(),
	}
}

func mapError(err error) error {
	switch status.Code(err) {
	case codes.ResourceExhausted:
		return quota.ErrInsufficientStorage

	case codes.NotFound:
		return quota.ErrUserNotFound

	case codes.FailedPrecondition:
		return quota.ErrInvalidStorageUsage

	case codes.Canceled:
		return context.Canceled

	case codes.DeadlineExceeded:
		return context.DeadlineExceeded

	default:
		return fmt.Errorf(
			"%w: %v",
			quota.ErrStorageUnavailable,
			err,
		)
	}
}
