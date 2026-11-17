package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"file-service/internal/application/folder"
	"file-service/internal/application/ping"
	"file-service/internal/domain/models"
	"file-service/internal/domain/repository"

	pb "proto/file"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeFolderRepository struct {
	folders map[string]*models.FolderMetadata
}

func newFakeFolderRepository() *fakeFolderRepository {
	return &fakeFolderRepository{
		folders: make(map[string]*models.FolderMetadata),
	}
}

func (r *fakeFolderRepository) Create(
	ctx context.Context,
	f *models.FolderMetadata,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.folders[f.ID] = f
	return nil
}

func (r *fakeFolderRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.FolderMetadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	f, ok := r.folders[id]
	if !ok {
		return nil, errors.New("folder not found")
	}

	return f, nil
}

func (r *fakeFolderRepository) Rename(
	ctx context.Context,
	id string,
	ownerID string,
	name string,
	updatedAt time.Time,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	folder, ok := r.folders[id]
	if !ok || folder.OwnerID != ownerID {
		return repository.ErrNotFound
	}

	folder.Name = name
	folder.UpdatedAt = updatedAt

	return nil
}

func newTestServer() (*Server, *fakeFolderRepository) {
	repo := newFakeFolderRepository()

	return NewServer(
		ping.NewHandler(5*time.Second),
		folder.NewCreateHandler(repo),
	), repo
}

func TestGRPCCreateRootFolder(t *testing.T) {
	server, repo := newTestServer()

	response, err := server.CreateFolder(
		context.Background(),
		&pb.CreateFolderRequest{
			Name:    "Documents",
			OwnerId: "user-1",
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	if response.GetId() == "" {
		t.Fatal("expected generated folder ID")
	}

	if response.GetName() != "Documents" {
		t.Fatalf("unexpected name: %s", response.GetName())
	}

	if response.GetOwnerId() != "user-1" {
		t.Fatal("unexpected owner ID")
	}

	if response.GetParentId() != "" {
		t.Fatal("root folder should not have a parent")
	}

	if response.GetCreatedAt() == "" {
		t.Fatal("expected creation timestamp")
	}

	if repo.folders[response.GetId()] == nil {
		t.Fatal("folder was not saved")
	}
}

func TestGRPCCreateNestedFolder(t *testing.T) {
	server, repo := newTestServer()

	parent, err := server.CreateFolder(
		context.Background(),
		&pb.CreateFolderRequest{
			Name:    "Documents",
			OwnerId: "user-1",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	child, err := server.CreateFolder(
		context.Background(),
		&pb.CreateFolderRequest{
			Name:     "Projects",
			ParentId: parent.GetId(),
			OwnerId:  "user-1",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if child.GetParentId() != parent.GetId() {
		t.Fatal("incorrect parent folder ID")
	}

	if repo.folders[child.GetId()] == nil {
		t.Fatal("child folder was not saved")
	}
}

func TestGRPCCreateFolderInvalidRequests(t *testing.T) {
	tests := []struct {
		name     string
		request  *pb.CreateFolderRequest
		wantCode codes.Code
	}{
		{
			name: "empty name",
			request: &pb.CreateFolderRequest{
				OwnerId: "user-1",
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "empty owner",
			request: &pb.CreateFolderRequest{
				Name: "Documents",
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "missing parent",
			request: &pb.CreateFolderRequest{
				Name:     "Child",
				ParentId: "missing-parent",
				OwnerId:  "user-1",
			},
			wantCode: codes.NotFound,
		},
		{
			name:     "nil request",
			request:  nil,
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, _ := newTestServer()

			_, err := server.CreateFolder(
				context.Background(),
				tt.request,
			)

			if status.Code(err) != tt.wantCode {
				t.Fatalf(
					"expected %v, got %v: %v",
					tt.wantCode,
					status.Code(err),
					err,
				)
			}
		})
	}
}

func TestGRPCCreateFolderDifferentOwner(t *testing.T) {
	server, repo := newTestServer()

	repo.folders["parent-1"] = &models.FolderMetadata{
		ID:      "parent-1",
		OwnerID: "user-2",
	}

	_, err := server.CreateFolder(
		context.Background(),
		&pb.CreateFolderRequest{
			Name:     "PrivateChild",
			ParentId: "parent-1",
			OwnerId:  "user-1",
		},
	)

	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func (r *fakeFolderRepository) ListByParentID(
	ctx context.Context,
	parentID string,
	ownerID string,
) ([]*models.FolderMetadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	folders := make([]*models.FolderMetadata, 0)

	for _, folder := range r.folders {
		if folder.ParentID == parentID && folder.OwnerID == ownerID {
			folders = append(folders, folder)
		}
	}

	return folders, nil
}
