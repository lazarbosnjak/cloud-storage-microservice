package folder

import (
	"context"
	"errors"
	"testing"
	"time"

	"file-service/internal/domain/models"
	"file-service/internal/domain/repository"
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
	folder *models.FolderMetadata,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.folders[folder.ID] = folder
	return nil
}

func (r *fakeFolderRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.FolderMetadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	folder, ok := r.folders[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return folder, nil
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

func TestCreateRootFolder(t *testing.T) {
	repo := newFakeFolderRepository()
	handler := NewCreateHandler(repo)

	result, err := handler.Handle(
		context.Background(),
		CreateCommand{
			Name:    "Documents",
			OwnerID: "user-1",
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	if result.ID == "" || result.Name != "Documents" {
		t.Fatal("invalid created folder")
	}

	if result.ParentID != "" {
		t.Fatal("root folder should not have a parent")
	}

	if repo.folders[result.ID] == nil {
		t.Fatal("folder was not saved")
	}
}

func TestCreateNestedFolders(t *testing.T) {
	repo := newFakeFolderRepository()
	handler := NewCreateHandler(repo)

	parentID := ""

	for i := 0; i < 20; i++ {
		result, err := handler.Handle(
			context.Background(),
			CreateCommand{
				Name:     "Nested",
				ParentID: parentID,
				OwnerID:  "user-1",
			},
		)
		if err != nil {
			t.Fatalf("depth %d: %v", i, err)
		}

		if result.ParentID != parentID {
			t.Fatalf("wrong parent at depth %d", i)
		}

		parentID = result.ID
	}

	if len(repo.folders) != 20 {
		t.Fatal("expected 20 folders")
	}
}

func TestCreateFolderInvalidParent(t *testing.T) {
	repo := newFakeFolderRepository()
	handler := NewCreateHandler(repo)

	_, err := handler.Handle(
		context.Background(),
		CreateCommand{
			Name:     "Child",
			ParentID: "missing-folder",
			OwnerID:  "user-1",
		},
	)

	if !errors.Is(err, ErrInvalidParent) {
		t.Fatalf("expected ErrInvalidParent, got %v", err)
	}
}

func TestCreateFolderDifferentOwner(t *testing.T) {
	repo := newFakeFolderRepository()
	repo.folders["parent-1"] = &models.FolderMetadata{
		ID:      "parent-1",
		OwnerID: "user-2",
	}

	handler := NewCreateHandler(repo)

	_, err := handler.Handle(
		context.Background(),
		CreateCommand{
			Name:     "Child",
			ParentID: "parent-1",
			OwnerID:  "user-1",
		},
	)

	if !errors.Is(err, ErrInvalidParent) {
		t.Fatalf("expected ErrInvalidParent, got %v", err)
	}
}

func TestCreateFolderInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		command CreateCommand
		wantErr error
	}{
		{
			name: "empty name",
			command: CreateCommand{
				OwnerID: "user-1",
			},
			wantErr: ErrInvalidName,
		},
		{
			name: "empty owner",
			command: CreateCommand{
				Name: "Documents",
			},
			wantErr: ErrInvalidOwner,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewCreateHandler(newFakeFolderRepository())

			_, err := handler.Handle(context.Background(), tt.command)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
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
