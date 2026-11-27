package file

import (
	"context"
	"errors"
	"testing"
	"time"

	"file-service/internal/domain/models"
	"file-service/internal/domain/repository"
)

type fakeFileRepository struct {
	files map[string]*models.FileMetadata
}

func newFakeFileRepository() *fakeFileRepository {
	return &fakeFileRepository{
		files: make(map[string]*models.FileMetadata),
	}
}

func (r *fakeFileRepository) Create(
	ctx context.Context,
	file *models.FileMetadata,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.files[file.ID] = file
	return nil
}

func (r *fakeFileRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.FileMetadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	file, ok := r.files[id]
	if !ok {
		return nil, repository.ErrNotFound
	}

	return file, nil
}

func (r *fakeFileRepository) Rename(
	ctx context.Context,
	id string,
	ownerID string,
	name string,
	updatedAt time.Time,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	file, ok := r.files[id]
	if !ok || file.OwnerID != ownerID {
		return repository.ErrNotFound
	}

	file.Name = name
	file.UpdatedAt = updatedAt

	return nil
}

func TestRenameFileSuccess(t *testing.T) {
	repo := newFakeFileRepository()

	createdAt := time.Now().UTC().Add(-time.Hour)

	repo.files["file-1"] = &models.FileMetadata{
		ID:          "file-1",
		Name:        "old.txt",
		Size:        1024,
		ContentType: "text/plain",
		StorageKey:  "storage/file-1",
		ParentID:    "folder-1",
		OwnerID:     "user-1",
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}

	handler := NewRenameHandler(repo)

	err := handler.Handle(context.Background(), RenameCommand{
		ID:      "file-1",
		OwnerID: "user-1",
		Name:    "new.txt",
	})
	if err != nil {
		t.Fatal(err)
	}

	saved := repo.files["file-1"]

	if saved.Name != "new.txt" {
		t.Fatalf("expected new.txt, got %s", saved.Name)
	}

	if saved.ID != "file-1" ||
		saved.Size != 1024 ||
		saved.ContentType != "text/plain" ||
		saved.StorageKey != "storage/file-1" ||
		saved.ParentID != "folder-1" ||
		saved.OwnerID != "user-1" ||
		!saved.CreatedAt.Equal(createdAt) {
		t.Fatal("rename changed unrelated file metadata")
	}

	if !saved.UpdatedAt.After(createdAt) {
		t.Fatal("UpdatedAt was not updated")
	}
}

func TestRenameFileInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		command RenameCommand
		wantErr error
	}{
		{
			name: "empty file ID",
			command: RenameCommand{
				OwnerID: "user-1",
				Name:    "new.txt",
			},
			wantErr: ErrInvalidFileID,
		},
		{
			name: "empty owner",
			command: RenameCommand{
				ID:   "file-1",
				Name: "new.txt",
			},
			wantErr: ErrInvalidOwner,
		},
		{
			name: "empty name",
			command: RenameCommand{
				ID:      "file-1",
				OwnerID: "user-1",
				Name:    " ",
			},
			wantErr: ErrInvalidName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewRenameHandler(newFakeFileRepository())

			err := handler.Handle(context.Background(), tt.command)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestRenameFileNotFound(t *testing.T) {
	handler := NewRenameHandler(newFakeFileRepository())

	err := handler.Handle(context.Background(), RenameCommand{
		ID:      "missing-file",
		OwnerID: "user-1",
		Name:    "new.txt",
	})

	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRenameFileDifferentOwner(t *testing.T) {
	repo := newFakeFileRepository()

	repo.files["file-1"] = &models.FileMetadata{
		ID:      "file-1",
		Name:    "private.txt",
		OwnerID: "user-2",
	}

	handler := NewRenameHandler(repo)

	err := handler.Handle(context.Background(), RenameCommand{
		ID:      "file-1",
		OwnerID: "user-1",
		Name:    "changed.txt",
	})

	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if repo.files["file-1"].Name != "private.txt" {
		t.Fatal("file belonging to another owner was modified")
	}
}

func TestRenameFileCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	handler := NewRenameHandler(newFakeFileRepository())

	err := handler.Handle(ctx, RenameCommand{
		ID:      "file-1",
		OwnerID: "user-1",
		Name:    "new.txt",
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func (r *fakeFileRepository) ListByParentID(
	ctx context.Context,
	parentID string,
	ownerID string,
) ([]*models.FileMetadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	files := make([]*models.FileMetadata, 0)

	for _, file := range r.files {
		if file.ParentID == parentID && file.OwnerID == ownerID {
			files = append(files, file)
		}
	}

	return files, nil
}
