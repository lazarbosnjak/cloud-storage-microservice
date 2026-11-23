package file

import (
	"context"
	"log/slog"

	appfile "api-gateway/internal/application/file"
	filePb "proto/file"
)

type GRPCClient struct {
	client filePb.FileClient
	logger *slog.Logger
}

func NewGRPCClient(client filePb.FileClient, logger *slog.Logger) *GRPCClient {
	return &GRPCClient{
		client: client,
		logger: logger,
	}
}

func (c *GRPCClient) GetFile(
	ctx context.Context,
	id string,
	ownerID string,
) (*appfile.FileMetadata, error) {
	reply, err := c.client.GetFile(ctx, &filePb.GetFileRequest{
		Id:      id,
		OwnerId: ownerID,
	})
	if err != nil {
		c.logger.Error("get file failed", "err", err)
		return nil, err
	}

	return &appfile.FileMetadata{
		ID:          reply.GetId(),
		Name:        reply.GetName(),
		Size:        reply.GetSize(),
		ContentType: reply.GetContentType(),
		ParentID:    reply.GetParentId(),
		OwnerID:     reply.GetOwnerId(),
		CreatedAt:   reply.GetCreatedAt(),
		UpdatedAt:   reply.GetUpdatedAt(),
	}, nil
}

func (c *GRPCClient) GetFolder(
	ctx context.Context,
	id string,
	ownerID string,
) (*appfile.FolderMetadata, error) {
	reply, err := c.client.GetFolder(ctx, &filePb.GetFolderRequest{
		Id:      id,
		OwnerId: ownerID,
	})
	if err != nil {
		c.logger.Error("get folder failed", "err", err)
		return nil, err
	}

	return &appfile.FolderMetadata{
		ID:        reply.GetId(),
		Name:      reply.GetName(),
		ParentID:  reply.GetParentId(),
		OwnerID:   reply.GetOwnerId(),
		CreatedAt: reply.GetCreatedAt(),
		UpdatedAt: reply.GetUpdatedAt(),
	}, nil
}
