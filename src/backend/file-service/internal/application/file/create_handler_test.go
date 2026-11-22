package file

import (
	"context"
	"errors"
	"testing"

	"file-service/internal/domain/models"
	"file-service/internal/domain/repository"
)

type createFileRepo struct {
	repository.FileRepository
	created *models.FileMetadata
	err     error
}

func (r *createFileRepo) Create(
	ctx context.Context,
	file *models.FileMetadata,
) error {
	r.created = file
	return r.err
}

type createFolderRepo struct {
	repository.FolderRepository
	parent *models.FolderMetadata
	err    error
}

func (r *createFolderRepo) GetByID(
	ctx context.Context,
	id string,
) (*models.FolderMetadata, error) {
	return r.parent, r.err
}

func validCreateCommand() CreateCommand {
	return CreateCommand{
		Name:        "document.pdf",
		Size:        1024,
		ContentType: "application/pdf",
		StorageKey:  "files/document.pdf",
		OwnerID:     "user-123",
	}
}

func TestCreateFileSuccess(t *testing.T) {
	files := &createFileRepo{}
	folders := &createFolderRepo{}

	handler := NewCreateHandler(files, folders)

	result, err := handler.Handle(
		context.Background(),
		validCreateCommand(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil || result.ID == "" {
		t.Fatal("expected generated file ID")
	}

	if files.created == nil {
		t.Fatal("expected repository Create to be called")
	}

	if result.Name != "document.pdf" {
		t.Fatalf("unexpected name: %s", result.Name)
	}

	if result.OwnerID != "user-123" {
		t.Fatalf("unexpected owner: %s", result.OwnerID)
	}

	if result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() {
		t.Fatal("expected timestamps")
	}
}

func TestCreateFileInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		change func(*CreateCommand)
		want   error
	}{
		{
			name: "empty name",
			change: func(c *CreateCommand) {
				c.Name = " "
			},
			want: ErrInvalidName,
		},
		{
			name: "empty owner",
			change: func(c *CreateCommand) {
				c.OwnerID = ""
			},
			want: ErrInvalidOwner,
		},
		{
			name: "zero size",
			change: func(c *CreateCommand) {
				c.Size = 0
			},
			want: ErrInvalidSize,
		},
		{
			name: "empty content type",
			change: func(c *CreateCommand) {
				c.ContentType = ""
			},
			want: ErrInvalidContentType,
		},
		{
			name: "empty storage key",
			change: func(c *CreateCommand) {
				c.StorageKey = ""
			},
			want: ErrInvalidStorageKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := validCreateCommand()
			tt.change(&cmd)

			files := &createFileRepo{}
			handler := NewCreateHandler(files, &createFolderRepo{})

			_, err := handler.Handle(context.Background(), cmd)

			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}

			if files.created != nil {
				t.Fatal("invalid file must not be saved")
			}
		})
	}
}

func TestCreateFileParentOwnership(t *testing.T) {
	cmd := validCreateCommand()
	cmd.ParentID = "folder-123"

	tests := []struct {
		name   string
		parent *models.FolderMetadata
		err    error
		want   error
	}{
		{
			name: "valid parent",
			parent: &models.FolderMetadata{
				ID:      "folder-123",
				OwnerID: "user-123",
			},
		},
		{
			name: "different owner",
			parent: &models.FolderMetadata{
				ID:      "folder-123",
				OwnerID: "another-user",
			},
			want: ErrInvalidParent,
		},
		{
			name: "missing parent",
			err:  repository.ErrNotFound,
			want: ErrInvalidParent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := &createFileRepo{}
			folders := &createFolderRepo{
				parent: tt.parent,
				err:    tt.err,
			}

			handler := NewCreateHandler(files, folders)
			_, err := handler.Handle(context.Background(), cmd)

			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}

			if tt.want == nil && files.created == nil {
				t.Fatal("expected file to be saved")
			}

			if tt.want != nil && files.created != nil {
				t.Fatal("file must not be saved")
			}
		})
	}
}

func TestCreateFileRepositoryError(t *testing.T) {
	files := &createFileRepo{
		err: errors.New("database unavailable"),
	}

	handler := NewCreateHandler(files, &createFolderRepo{})

	_, err := handler.Handle(
		context.Background(),
		validCreateCommand(),
	)

	if err == nil {
		t.Fatal("expected repository error")
	}
}

func TestCreateFileCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	files := &createFileRepo{}
	handler := NewCreateHandler(files, &createFolderRepo{})

	_, err := handler.Handle(ctx, validCreateCommand())

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled context, got %v", err)
	}

	if files.created != nil {
		t.Fatal("file must not be saved")
	}
}
