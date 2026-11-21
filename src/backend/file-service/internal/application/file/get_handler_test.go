package file

import (
	"context"
	"errors"
	"testing"
	"time"

	"file-service/internal/domain/models"
	"file-service/internal/domain/repository"
)

type fakeGetFileRepository struct {
	files map[string]*models.FileMetadata
	err   error
}

func (r *fakeGetFileRepository) Create(
	ctx context.Context,
	value *models.FileMetadata,
) error {
	return nil
}

func (r *fakeGetFileRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.FileMetadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if r.err != nil {
		return nil, r.err
	}

	value, ok := r.files[id]
	if !ok {
		return nil, repository.ErrNotFound
	}

	return value, nil
}

func (r *fakeGetFileRepository) Rename(
	ctx context.Context,
	id, ownerID, name string,
	updatedAt time.Time,
) error {
	return nil
}

func (r *fakeGetFileRepository) ListByParentID(
	ctx context.Context,
	parentID, ownerID string,
) ([]*models.FileMetadata, error) {
	return nil, nil
}

func TestGetFileSuccess(t *testing.T) {
	repo := &fakeGetFileRepository{
		files: map[string]*models.FileMetadata{
			"file-1": {
				ID:      "file-1",
				Name:    "report.pdf",
				OwnerID: "user-1",
				Size:    1024,
			},
		},
	}

	handler := NewGetHandler(repo)

	result, err := handler.Handle(
		context.Background(),
		GetQuery{
			ID:      "file-1",
			OwnerID: "user-1",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.ID != "file-1" || result.Name != "report.pdf" {
		t.Fatalf("unexpected file metadata: %+v", result)
	}

	if result.Size != 1024 {
		t.Fatalf("unexpected size: %d", result.Size)
	}
}

func TestGetFileNotFound(t *testing.T) {
	handler := NewGetHandler(&fakeGetFileRepository{
		files: map[string]*models.FileMetadata{},
	})

	_, err := handler.Handle(
		context.Background(),
		GetQuery{ID: "missing", OwnerID: "user-1"},
	)

	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetFileDifferentOwner(t *testing.T) {
	handler := NewGetHandler(&fakeGetFileRepository{
		files: map[string]*models.FileMetadata{
			"file-1": {
				ID:      "file-1",
				OwnerID: "user-2",
			},
		},
	})

	_, err := handler.Handle(
		context.Background(),
		GetQuery{ID: "file-1", OwnerID: "user-1"},
	)

	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetFileInvalidArguments(t *testing.T) {
	handler := NewGetHandler(&fakeGetFileRepository{})

	tests := []struct {
		name  string
		query GetQuery
		want  error
	}{
		{
			name:  "empty ID",
			query: GetQuery{OwnerID: "user-1"},
			want:  ErrInvalidFileID,
		},
		{
			name:  "empty owner",
			query: GetQuery{ID: "file-1"},
			want:  ErrInvalidOwner,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Handle(context.Background(), tt.query)
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
		})
	}
}

func TestGetFileCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	handler := NewGetHandler(&fakeGetFileRepository{})

	_, err := handler.Handle(
		ctx,
		GetQuery{ID: "file-1", OwnerID: "user-1"},
	)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestGetFileRepositoryError(t *testing.T) {
	dbErr := errors.New("database unavailable")

	handler := NewGetHandler(&fakeGetFileRepository{
		err: dbErr,
	})

	_, err := handler.Handle(
		context.Background(),
		GetQuery{ID: "file-1", OwnerID: "user-1"},
	)

	if !errors.Is(err, dbErr) {
		t.Fatalf("expected wrapped database error, got %v", err)
	}
}
