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

type CreateFileInput struct {
	Name        string
	Size        uint64
	ContentType string
	ParentID    string
	OwnerID     string
}

type CreatedFile struct {
	FileMetadata
	StorageKey string
}

type CreateFolderInput struct {
	Name     string
	ParentID string
	OwnerID  string
}

type CreatedFolder struct {
	ID        string
	Name      string
	ParentID  string
	OwnerID   string
	CreatedAt string
}

type FolderContents struct {
	Files   []*FileMetadata
	Folders []*FolderMetadata
}

type Client interface {
	GetFile(ctx context.Context, id, ownerID string) (*FileMetadata, error)
	GetFolder(ctx context.Context, id, ownerID string) (*FolderMetadata, error)

	CreateFile(ctx context.Context, input CreateFileInput) (*CreatedFile, error)
	CreateFolder(ctx context.Context, input CreateFolderInput) (*CreatedFolder, error)

	RenameFile(ctx context.Context, id, ownerID, name string) error
	RenameFolder(ctx context.Context, id, ownerID, name string) error

	ListFolderContents(ctx context.Context, folderID, ownerID string) (*FolderContents, error)
}
