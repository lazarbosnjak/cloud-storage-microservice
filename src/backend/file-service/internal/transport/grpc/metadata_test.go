package grpc

import (
	"context"
	"testing"
	"time"

	"file-service/internal/application/file"
	"file-service/internal/application/folder"
	"file-service/internal/domain/models"
	"file-service/internal/domain/repository"

	pb "proto/file"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeGRPCFileRepository struct {
	files map[string]*models.FileMetadata
}

func newFakeGRPCFileRepository() *fakeGRPCFileRepository {
	return &fakeGRPCFileRepository{
		files: make(map[string]*models.FileMetadata),
	}
}

func (r *fakeGRPCFileRepository) Create(
	ctx context.Context,
	value *models.FileMetadata,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.files[value.ID] = value
	return nil
}

func (r *fakeGRPCFileRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.FileMetadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	value, ok := r.files[id]
	if !ok {
		return nil, repository.ErrNotFound
	}

	return value, nil
}

func (r *fakeGRPCFileRepository) Rename(
	ctx context.Context,
	id, ownerID, name string,
	updatedAt time.Time,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	value, ok := r.files[id]
	if !ok || value.OwnerID != ownerID {
		return repository.ErrNotFound
	}

	value.Name = name
	value.UpdatedAt = updatedAt
	return nil
}

func (r *fakeGRPCFileRepository) ListByParentID(
	ctx context.Context,
	parentID, ownerID string,
) ([]*models.FileMetadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := make([]*models.FileMetadata, 0)

	for _, value := range r.files {
		if value.ParentID == parentID && value.OwnerID == ownerID {
			result = append(result, value)
		}
	}

	return result, nil
}

func TestGRPCRenameFile(t *testing.T) {
	server, _ := newTestServer()

	serverFile := &models.FileMetadata{
		ID:      "file-1",
		Name:    "old.txt",
		OwnerID: "user-1",
	}

	// Test server uses its own repository.
	// Insert through the GetFile handler's repository by
	// constructing a server with the shared fake repository.
	fileRepo := newFakeGRPCFileRepository()
	fileRepo.files[serverFile.ID] = serverFile

	folderRepo := newFakeFolderRepository()

	server = NewServer(
		nil,
		nil,
		file.NewRenameHandler(fileRepo),
		folder.NewRenameHandler(folderRepo),
		folder.NewListContentsHandler(folderRepo, fileRepo),
		file.NewGetHandler(fileRepo),
		folder.NewGetHandler(folderRepo),
	)

	_, err := server.RenameFile(
		context.Background(),
		&pb.RenameFileRequest{
			Id:      "file-1",
			OwnerId: "user-1",
			Name:    "new.txt",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if serverFile.Name != "new.txt" {
		t.Fatalf("file was not renamed: %s", serverFile.Name)
	}

	if serverFile.UpdatedAt.IsZero() {
		t.Fatal("expected updated timestamp")
	}
}

func TestGRPCRenameFolder(t *testing.T) {
	server, folderRepo := newTestServer()

	folderRepo.folders["folder-1"] = &models.FolderMetadata{
		ID:      "folder-1",
		Name:    "Old",
		OwnerID: "user-1",
	}

	_, err := server.RenameFolder(
		context.Background(),
		&pb.RenameFolderRequest{
			Id:      "folder-1",
			OwnerId: "user-1",
			Name:    "New",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if folderRepo.folders["folder-1"].Name != "New" {
		t.Fatal("folder was not renamed")
	}
}

func TestGRPCListFolderContents(t *testing.T) {
	folderRepo := newFakeFolderRepository()
	fileRepo := newFakeGRPCFileRepository()

	folderRepo.folders["parent"] = &models.FolderMetadata{
		ID:      "parent",
		OwnerID: "user-1",
	}

	folderRepo.folders["child"] = &models.FolderMetadata{
		ID:       "child",
		Name:     "Projects",
		ParentID: "parent",
		OwnerID:  "user-1",
	}

	fileRepo.files["file-1"] = &models.FileMetadata{
		ID:       "file-1",
		Name:     "report.pdf",
		ParentID: "parent",
		OwnerID:  "user-1",
		Size:     1024,
	}

	server := NewServer(
		nil,
		nil,
		file.NewRenameHandler(fileRepo),
		folder.NewRenameHandler(folderRepo),
		folder.NewListContentsHandler(folderRepo, fileRepo),
		file.NewGetHandler(fileRepo),
		folder.NewGetHandler(folderRepo),
	)

	reply, err := server.ListFolderContents(
		context.Background(),
		&pb.ListFolderContentsRequest{
			FolderId: "parent",
			OwnerId:  "user-1",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(reply.Folders) != 1 {
		t.Fatalf("expected 1 folder, got %d", len(reply.Folders))
	}

	if len(reply.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(reply.Files))
	}

	if reply.Files[0].GetName() != "report.pdf" {
		t.Fatal("unexpected file metadata")
	}
}

func TestGRPCGetFile(t *testing.T) {
	fileRepo := newFakeGRPCFileRepository()
	folderRepo := newFakeFolderRepository()

	fileRepo.files["file-1"] = &models.FileMetadata{
		ID:      "file-1",
		Name:    "report.pdf",
		OwnerID: "user-1",
		Size:    2048,
	}

	server := NewServer(
		nil,
		nil,
		file.NewRenameHandler(fileRepo),
		folder.NewRenameHandler(folderRepo),
		folder.NewListContentsHandler(folderRepo, fileRepo),
		file.NewGetHandler(fileRepo),
		folder.NewGetHandler(folderRepo),
	)

	reply, err := server.GetFile(
		context.Background(),
		&pb.GetFileRequest{
			Id:      "file-1",
			OwnerId: "user-1",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if reply.GetName() != "report.pdf" || reply.GetSize() != 2048 {
		t.Fatalf("unexpected metadata: %+v", reply)
	}
}

func TestGRPCGetFolder(t *testing.T) {
	server, folderRepo := newTestServer()

	folderRepo.folders["folder-1"] = &models.FolderMetadata{
		ID:      "folder-1",
		Name:    "Documents",
		OwnerID: "user-1",
	}

	reply, err := server.GetFolder(
		context.Background(),
		&pb.GetFolderRequest{
			Id:      "folder-1",
			OwnerId: "user-1",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if reply.GetName() != "Documents" {
		t.Fatalf("unexpected folder name: %s", reply.GetName())
	}
}

func TestGRPCMetadataInvalidRequests(t *testing.T) {
	server, _ := newTestServer()

	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "nil RenameFile",
			call: func() error {
				_, err := server.RenameFile(context.Background(), nil)
				return err
			},
		},
		{
			name: "nil RenameFolder",
			call: func() error {
				_, err := server.RenameFolder(context.Background(), nil)
				return err
			},
		},
		{
			name: "nil ListFolderContents",
			call: func() error {
				_, err := server.ListFolderContents(context.Background(), nil)
				return err
			},
		},
		{
			name: "nil GetFile",
			call: func() error {
				_, err := server.GetFile(context.Background(), nil)
				return err
			},
		},
		{
			name: "nil GetFolder",
			call: func() error {
				_, err := server.GetFolder(context.Background(), nil)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if code := status.Code(tt.call()); code != codes.InvalidArgument {
				t.Fatalf("expected InvalidArgument, got %v", code)
			}
		})
	}
}

func TestGRPCMetadataNotFound(t *testing.T) {
	server, _ := newTestServer()

	_, err := server.GetFolder(
		context.Background(),
		&pb.GetFolderRequest{
			Id:      "missing",
			OwnerId: "user-1",
		},
	)

	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}
