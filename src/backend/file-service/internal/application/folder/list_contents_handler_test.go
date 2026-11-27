package folder

import (
	"context"
	"errors"
	"testing"

	"file-service/internal/domain/models"
	"file-service/internal/domain/repository"
	"time"
)

type fakeListFileRepository struct {
	files map[string]*models.FileMetadata
}

func newFakeListFileRepository() *fakeListFileRepository {
	return &fakeListFileRepository{
		files: make(map[string]*models.FileMetadata),
	}
}

func (r *fakeListFileRepository) Create(
	ctx context.Context,
	file *models.FileMetadata,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.files[file.ID] = file
	return nil
}

func (r *fakeListFileRepository) GetByID(
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

func (r *fakeListFileRepository) Rename(
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

func (r *fakeListFileRepository) ListByParentID(
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

func TestListFolderContentsSuccess(t *testing.T) {
	folderRepo := newFakeFolderRepository()
	fileRepo := newFakeListFileRepository()

	folderRepo.folders["parent-1"] = &models.FolderMetadata{
		ID:      "parent-1",
		Name:    "Dokumenti",
		OwnerID: "user-1",
	}

	folderRepo.folders["child-1"] = &models.FolderMetadata{
		ID:       "child-1",
		Name:     "Fakultet",
		ParentID: "parent-1",
		OwnerID:  "user-1",
	}

	fileRepo.files["file-1"] = &models.FileMetadata{
		ID:       "file-1",
		Name:     "projekat.pdf",
		ParentID: "parent-1",
		OwnerID:  "user-1",
	}

	fileRepo.files["file-2"] = &models.FileMetadata{
		ID:       "file-2",
		Name:     "private.txt",
		ParentID: "parent-1",
		OwnerID:  "user-2",
	}

	handler := NewListContentsHandler(folderRepo, fileRepo)

	result, err := handler.Handle(context.Background(), ListContentsQuery{
		FolderID: "parent-1",
		OwnerID:  "user-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Folders) != 1 {
		t.Fatalf("expected 1 folder, got %d", len(result.Folders))
	}

	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}

	if result.Folders[0].Name != "Fakultet" {
		t.Fatal("unexpected folder")
	}

	if result.Files[0].Name != "projekat.pdf" {
		t.Fatal("unexpected file")
	}
}

func TestListRootContents(t *testing.T) {
	folderRepo := newFakeFolderRepository()
	fileRepo := newFakeListFileRepository()

	folderRepo.folders["root-folder"] = &models.FolderMetadata{
		ID:      "root-folder",
		Name:    "Dokumenti",
		OwnerID: "user-1",
	}

	fileRepo.files["root-file"] = &models.FileMetadata{
		ID:      "root-file",
		Name:    "readme.txt",
		OwnerID: "user-1",
	}

	handler := NewListContentsHandler(folderRepo, fileRepo)

	result, err := handler.Handle(context.Background(), ListContentsQuery{
		OwnerID: "user-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Folders) != 1 || len(result.Files) != 1 {
		t.Fatalf(
			"expected 1 folder and 1 file, got %d folders and %d files",
			len(result.Folders),
			len(result.Files),
		)
	}
}

func TestListFolderContentsNotFound(t *testing.T) {
	handler := NewListContentsHandler(
		newFakeFolderRepository(),
		newFakeListFileRepository(),
	)

	_, err := handler.Handle(context.Background(), ListContentsQuery{
		FolderID: "missing-folder",
		OwnerID:  "user-1",
	})

	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListFolderContentsDifferentOwner(t *testing.T) {
	folderRepo := newFakeFolderRepository()

	folderRepo.folders["folder-1"] = &models.FolderMetadata{
		ID:      "folder-1",
		OwnerID: "user-2",
	}

	handler := NewListContentsHandler(
		folderRepo,
		newFakeListFileRepository(),
	)

	_, err := handler.Handle(context.Background(), ListContentsQuery{
		FolderID: "folder-1",
		OwnerID:  "user-1",
	})

	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListFolderContentsInvalidOwner(t *testing.T) {
	handler := NewListContentsHandler(
		newFakeFolderRepository(),
		newFakeListFileRepository(),
	)

	_, err := handler.Handle(context.Background(), ListContentsQuery{
		FolderID: "folder-1",
		OwnerID:  " ",
	})

	if !errors.Is(err, ErrInvalidOwner) {
		t.Fatalf("expected ErrInvalidOwner, got %v", err)
	}
}

func TestListFolderContentsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	handler := NewListContentsHandler(
		newFakeFolderRepository(),
		newFakeListFileRepository(),
	)

	_, err := handler.Handle(ctx, ListContentsQuery{
		OwnerID: "user-1",
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
