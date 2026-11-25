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

func (c *GRPCClient) CreateFile(
	ctx context.Context,
	input appfile.CreateFileInput,
) (*appfile.CreatedFile, error) {
	reply, err := c.client.CreateFile(ctx, &filePb.CreateFileRequest{
		Name:        input.Name,
		Size:        input.Size,
		ContentType: input.ContentType,
		StorageKey:  input.StorageKey,
		ParentId:    input.ParentID,
		OwnerId:     input.OwnerID,
	})
	if err != nil {
		c.logger.Error("create file failed", "err", err)
		return nil, err
	}

	return &appfile.CreatedFile{
		FileMetadata: appfile.FileMetadata{
			ID:          reply.GetId(),
			Name:        reply.GetName(),
			Size:        reply.GetSize(),
			ContentType: reply.GetContentType(),
			ParentID:    reply.GetParentId(),
			OwnerID:     reply.GetOwnerId(),
			CreatedAt:   reply.GetCreatedAt(),
			UpdatedAt:   reply.GetUpdatedAt(),
		},
		StorageKey: reply.GetStorageKey(),
	}, nil
}

func (c *GRPCClient) CreateFolder(
	ctx context.Context,
	input appfile.CreateFolderInput,
) (*appfile.CreatedFolder, error) {
	reply, err := c.client.CreateFolder(ctx, &filePb.CreateFolderRequest{
		Name:     input.Name,
		ParentId: input.ParentID,
		OwnerId:  input.OwnerID,
	})
	if err != nil {
		c.logger.Error("create folder failed", "err", err)
		return nil, err
	}

	return &appfile.CreatedFolder{
		ID:        reply.GetId(),
		Name:      reply.GetName(),
		ParentID:  reply.GetParentId(),
		OwnerID:   reply.GetOwnerId(),
		CreatedAt: reply.GetCreatedAt(),
	}, nil
}

func (c *GRPCClient) RenameFile(
	ctx context.Context,
	id, ownerID, name string,
) error {
	_, err := c.client.RenameFile(ctx, &filePb.RenameFileRequest{
		Id:      id,
		OwnerId: ownerID,
		Name:    name,
	})
	if err != nil {
		c.logger.Error("rename file failed", "err", err)
	}
	return err
}

func (c *GRPCClient) RenameFolder(
	ctx context.Context,
	id, ownerID, name string,
) error {
	_, err := c.client.RenameFolder(ctx, &filePb.RenameFolderRequest{
		Id:      id,
		OwnerId: ownerID,
		Name:    name,
	})
	if err != nil {
		c.logger.Error("rename folder failed", "err", err)
	}
	return err
}

func (c *GRPCClient) ListFolderContents(
	ctx context.Context,
	folderID, ownerID string,
) (*appfile.FolderContents, error) {
	reply, err := c.client.ListFolderContents(
		ctx,
		&filePb.ListFolderContentsRequest{
			FolderId: folderID,
			OwnerId:  ownerID,
		},
	)
	if err != nil {
		c.logger.Error("list folder contents failed", "err", err)
		return nil, err
	}

	result := &appfile.FolderContents{
		Files:   make([]*appfile.FileMetadata, 0, len(reply.GetFiles())),
		Folders: make([]*appfile.FolderMetadata, 0, len(reply.GetFolders())),
	}

	for _, f := range reply.GetFiles() {
		result.Files = append(result.Files, &appfile.FileMetadata{
			ID:          f.GetId(),
			Name:        f.GetName(),
			Size:        f.GetSize(),
			ContentType: f.GetContentType(),
			ParentID:    f.GetParentId(),
			OwnerID:     f.GetOwnerId(),
			CreatedAt:   f.GetCreatedAt(),
			UpdatedAt:   f.GetUpdatedAt(),
		})
	}

	for _, f := range reply.GetFolders() {
		result.Folders = append(result.Folders, &appfile.FolderMetadata{
			ID:        f.GetId(),
			Name:      f.GetName(),
			ParentID:  f.GetParentId(),
			OwnerID:   f.GetOwnerId(),
			CreatedAt: f.GetCreatedAt(),
			UpdatedAt: f.GetUpdatedAt(),
		})
	}

	return result, nil
}
