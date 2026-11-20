package folder

import (
	"context"
	"errors"
	"testing"

	"file-service/internal/domain/models"
	"file-service/internal/domain/repository"
)

func TestGetFolderSuccess(t *testing.T) {
	repo := newFakeFolderRepository()

	repo.folders["folder-1"] = &models.FolderMetadata{
		ID:       "folder-1",
		Name:     "Documents",
		OwnerID:  "user-1",
		ParentID: "",
	}

	handler := NewGetHandler(repo)

	result, err := handler.Handle(
		context.Background(),
		GetQuery{
			ID:      "folder-1",
			OwnerID: "user-1",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.ID != "folder-1" || result.Name != "Documents" {
		t.Fatalf("unexpected folder metadata: %+v", result)
	}
}

func TestGetFolderNotFound(t *testing.T) {
	handler := NewGetHandler(newFakeFolderRepository())

	_, err := handler.Handle(
		context.Background(),
		GetQuery{ID: "missing", OwnerID: "user-1"},
	)

	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetFolderDifferentOwner(t *testing.T) {
	repo := newFakeFolderRepository()

	repo.folders["folder-1"] = &models.FolderMetadata{
		ID:      "folder-1",
		OwnerID: "user-2",
	}

	handler := NewGetHandler(repo)

	_, err := handler.Handle(
		context.Background(),
		GetQuery{ID: "folder-1", OwnerID: "user-1"},
	)

	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetFolderInvalidArguments(t *testing.T) {
	handler := NewGetHandler(newFakeFolderRepository())

	tests := []struct {
		name  string
		query GetQuery
		want  error
	}{
		{
			name:  "empty ID",
			query: GetQuery{OwnerID: "user-1"},
			want:  ErrInvalidFolderID,
		},
		{
			name:  "empty owner",
			query: GetQuery{ID: "folder-1"},
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

func TestGetFolderCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	handler := NewGetHandler(newFakeFolderRepository())

	_, err := handler.Handle(
		ctx,
		GetQuery{ID: "folder-1", OwnerID: "user-1"},
	)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
