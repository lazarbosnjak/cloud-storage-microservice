package folder

import (
	"context"
	"errors"
	"testing"
	"time"

	"file-service/internal/domain/models"
	"file-service/internal/domain/repository"
)

func TestRenameFolderSuccess(t *testing.T) {
	repo := newFakeFolderRepository()

	createdAt := time.Now().UTC().Add(-time.Hour)

	repo.folders["folder-1"] = &models.FolderMetadata{
		ID:        "folder-1",
		Name:      "Old Name",
		ParentID:  "parent-1",
		OwnerID:   "user-1",
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	handler := NewRenameHandler(repo)

	err := handler.Handle(context.Background(), RenameCommand{
		ID:      "folder-1",
		OwnerID: "user-1",
		Name:    "New Name",
	})
	if err != nil {
		t.Fatal(err)
	}

	saved := repo.folders["folder-1"]

	if saved.Name != "New Name" {
		t.Fatalf("expected New Name, got %s", saved.Name)
	}

	if saved.ID != "folder-1" ||
		saved.ParentID != "parent-1" ||
		saved.OwnerID != "user-1" ||
		!saved.CreatedAt.Equal(createdAt) {
		t.Fatal("rename changed unrelated folder metadata")
	}

	if !saved.UpdatedAt.After(createdAt) {
		t.Fatal("UpdatedAt was not updated")
	}
}

func TestRenameFolderInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		command RenameCommand
		wantErr error
	}{
		{
			name: "empty folder ID",
			command: RenameCommand{
				OwnerID: "user-1",
				Name:    "New Name",
			},
			wantErr: ErrInvalidFolderID,
		},
		{
			name: "empty owner",
			command: RenameCommand{
				ID:   "folder-1",
				Name: "New Name",
			},
			wantErr: ErrInvalidOwner,
		},
		{
			name: "empty name",
			command: RenameCommand{
				ID:      "folder-1",
				OwnerID: "user-1",
				Name:    " ",
			},
			wantErr: ErrInvalidName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewRenameHandler(newFakeFolderRepository())

			err := handler.Handle(context.Background(), tt.command)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestRenameFolderNotFound(t *testing.T) {
	handler := NewRenameHandler(newFakeFolderRepository())

	err := handler.Handle(context.Background(), RenameCommand{
		ID:      "missing-folder",
		OwnerID: "user-1",
		Name:    "New Name",
	})

	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRenameFolderDifferentOwner(t *testing.T) {
	repo := newFakeFolderRepository()

	repo.folders["folder-1"] = &models.FolderMetadata{
		ID:      "folder-1",
		Name:    "Private",
		OwnerID: "user-2",
	}

	handler := NewRenameHandler(repo)

	err := handler.Handle(context.Background(), RenameCommand{
		ID:      "folder-1",
		OwnerID: "user-1",
		Name:    "Changed",
	})

	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if repo.folders["folder-1"].Name != "Private" {
		t.Fatal("folder belonging to another owner was modified")
	}
}

func TestRenameFolderCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	handler := NewRenameHandler(newFakeFolderRepository())

	err := handler.Handle(ctx, RenameCommand{
		ID:      "folder-1",
		OwnerID: "user-1",
		Name:    "New Name",
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
