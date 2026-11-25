package file

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"file-service/internal/domain/models"
	"file-service/internal/domain/quota"
	"file-service/internal/domain/repository"
)

type uploadFileRepo struct {
	repository.UploadRepository
	existing   *models.FileMetadata
	created    *models.FileMetadata
	updated    *models.FileMetadata
	findErr    error
	createErr  error
	updateErr  error
	expectedAt time.Time
}

func (r *uploadFileRepo) GetByNameAndParent(
	ctx context.Context,
	name string,
	parentID string,
	ownerID string,
) (*models.FileMetadata, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}

	if r.existing == nil {
		return nil, repository.ErrNotFound
	}

	return r.existing, nil
}

func (r *uploadFileRepo) Create(
	ctx context.Context,
	file *models.FileMetadata,
) error {
	if r.createErr != nil {
		return r.createErr
	}

	r.created = file
	return nil
}

func (r *uploadFileRepo) UpdateVersion(
	ctx context.Context,
	file *models.FileMetadata,
	expectedUpdatedAt time.Time,
) error {
	if r.updateErr != nil {
		return r.updateErr
	}

	r.updated = file
	r.expectedAt = expectedUpdatedAt

	return nil
}

type uploadQuotaClient struct {
	info   quota.StorageInfo
	deltas []int64
	err    error
}

func (c *uploadQuotaClient) GetStorage(
	ctx context.Context,
	userID string,
) (quota.StorageInfo, error) {
	return c.info, c.err
}

func (c *uploadQuotaClient) AdjustStorageUsage(
	ctx context.Context,
	userID string,
	delta int64,
) (quota.StorageInfo, error) {
	c.deltas = append(c.deltas, delta)

	if c.err != nil {
		return quota.StorageInfo{}, c.err
	}

	c.info.Used += delta
	c.info.Available = c.info.Quota - c.info.Used

	return c.info, nil
}

func validUploadCommand() UploadCommand {
	return UploadCommand{
		Name:        "document.pdf",
		Size:        1024,
		ContentType: "application/pdf",
		OwnerID:     "user-123",
	}
}

func TestUploadNewFile(t *testing.T) {
	files := &uploadFileRepo{}
	client := &uploadQuotaClient{}

	handler := NewUploadHandler(
		files,
		&createFolderRepo{},
		client,
	)

	cmd := validUploadCommand()

	result, err := handler.Handle(
		context.Background(),
		cmd,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID == "" || files.created == nil {
		t.Fatal("expected new file metadata")
	}

	if result.Name != cmd.Name ||
		result.Size != cmd.Size ||
		result.OwnerID != cmd.OwnerID ||
		result.ContentType != cmd.ContentType {
		t.Fatal("unexpected file metadata")
	}

	if !strings.HasPrefix(
		result.StorageKey,
		"placeholder/",
	) {
		t.Fatal("expected placeholder storage key")
	}

	if !reflect.DeepEqual(
		client.deltas,
		[]int64{1024},
	) {
		t.Fatalf("unexpected deltas: %v", client.deltas)
	}
}

func TestUploadExistingFile(t *testing.T) {
	tests := []struct {
		name  string
		size  uint64
		delta int64
	}{
		{"larger version", 2048, 1024},
		{"smaller version", 512, -512},
		{"same size", 1024, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now().UTC().
				Add(-time.Hour).
				Truncate(time.Millisecond)

			existing := &models.FileMetadata{
				ID:          "existing-file",
				Name:        "document.pdf",
				Size:        1024,
				ContentType: "application/pdf",
				OwnerID:     "user-123",
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			files := &uploadFileRepo{
				existing: existing,
			}

			client := &uploadQuotaClient{}

			handler := NewUploadHandler(
				files,
				&createFolderRepo{},
				client,
			)

			cmd := validUploadCommand()
			cmd.Size = tt.size

			result, err := handler.Handle(
				context.Background(),
				cmd,
			)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if files.created != nil ||
				files.updated == nil {
				t.Fatal("expected existing file update")
			}

			if result.ID != existing.ID {
				t.Fatal("file ID must remain unchanged")
			}

			if result.Size != tt.size {
				t.Fatal("unexpected file size")
			}

			if !result.CreatedAt.Equal(
				existing.CreatedAt,
			) {
				t.Fatal("CreatedAt must remain unchanged")
			}

			if !result.UpdatedAt.After(
				existing.UpdatedAt,
			) {
				t.Fatal("UpdatedAt must be newer")
			}

			if !files.expectedAt.Equal(
				existing.UpdatedAt,
			) {
				t.Fatal("unexpected version timestamp")
			}

			if !reflect.DeepEqual(
				client.deltas,
				[]int64{tt.delta},
			) {
				t.Fatalf(
					"unexpected deltas: %v",
					client.deltas,
				)
			}
		})
	}
}

func TestUploadInsufficientStorage(t *testing.T) {
	files := &uploadFileRepo{}

	client := &uploadQuotaClient{
		err: quota.ErrInsufficientStorage,
	}

	handler := NewUploadHandler(
		files,
		&createFolderRepo{},
		client,
	)

	_, err := handler.Handle(
		context.Background(),
		validUploadCommand(),
	)

	if !errors.Is(err, quota.ErrInsufficientStorage) {
		t.Fatalf("unexpected error: %v", err)
	}

	if files.created != nil {
		t.Fatal("file must not be saved")
	}
}

func TestUploadRollbackOnCreateError(t *testing.T) {
	dbErr := errors.New("database unavailable")

	files := &uploadFileRepo{
		createErr: dbErr,
	}

	client := &uploadQuotaClient{}

	handler := NewUploadHandler(
		files,
		&createFolderRepo{},
		client,
	)

	_, err := handler.Handle(
		context.Background(),
		validUploadCommand(),
	)

	if !errors.Is(err, dbErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(
		client.deltas,
		[]int64{1024, -1024},
	) {
		t.Fatalf("expected rollback, got %v", client.deltas)
	}
}

func TestUploadInvalidParent(t *testing.T) {
	files := &uploadFileRepo{}
	client := &uploadQuotaClient{}

	folders := &createFolderRepo{
		parent: &models.FolderMetadata{
			ID:      "folder-123",
			OwnerID: "another-user",
		},
	}

	handler := NewUploadHandler(
		files,
		folders,
		client,
	)

	cmd := validUploadCommand()
	cmd.ParentID = "folder-123"

	_, err := handler.Handle(
		context.Background(),
		cmd,
	)

	if !errors.Is(err, ErrInvalidParent) {
		t.Fatalf("unexpected error: %v", err)
	}

	if files.created != nil || len(client.deltas) != 0 {
		t.Fatal("unauthorized upload must not change storage")
	}
}

func TestUploadInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		change func(*UploadCommand)
		want   error
	}{
		{
			"empty name",
			func(c *UploadCommand) { c.Name = "" },
			ErrInvalidName,
		},
		{
			"empty owner",
			func(c *UploadCommand) { c.OwnerID = "" },
			ErrInvalidOwner,
		},
		{
			"zero size",
			func(c *UploadCommand) { c.Size = 0 },
			ErrInvalidSize,
		},
		{
			"size overflow",
			func(c *UploadCommand) { c.Size = ^uint64(0) },
			ErrInvalidSize,
		},
		{
			"empty content type",
			func(c *UploadCommand) { c.ContentType = "" },
			ErrInvalidContentType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := &uploadFileRepo{}
			client := &uploadQuotaClient{}

			handler := NewUploadHandler(
				files,
				&createFolderRepo{},
				client,
			)

			cmd := validUploadCommand()
			tt.change(&cmd)

			_, err := handler.Handle(
				context.Background(),
				cmd,
			)

			if !errors.Is(err, tt.want) {
				t.Fatalf(
					"expected %v, got %v",
					tt.want,
					err,
				)
			}

			if files.created != nil ||
				len(client.deltas) != 0 {
				t.Fatal("invalid upload changed storage")
			}
		})
	}
}
