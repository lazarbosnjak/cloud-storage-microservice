package folder

import (
	"context"
	"errors"
	"testing"

	"file-service/internal/domain/models"
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
		return nil, errors.New("folder not found")
	}
	return folder, nil
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
