package file

import "context"

type FileMetadata struct {
	ID          string
	Name        string
	Size        uint64
	ContentType string
	ParentID    string
	OwnerID     string
	CreatedAt   string
	UpdatedAt   string
}

type FolderMetadata struct {
	ID        string
	Name      string
	ParentID  string
	OwnerID   string
	CreatedAt string
	UpdatedAt string
}

type Client interface {
	GetFile(ctx context.Context, id, ownerID string) (*FileMetadata, error)
	GetFolder(ctx context.Context, id, ownerID string) (*FolderMetadata, error)
}
