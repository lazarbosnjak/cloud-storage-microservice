package file

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"file-service/internal/domain/models"
	"file-service/internal/domain/quota"
	"file-service/internal/domain/repository"
)

type UploadHandler struct {
	fileRepository   repository.UploadRepository
	folderRepository repository.FolderRepository
	quotaClient      quota.Client
	mu               sync.Mutex
}

func NewUploadHandler(
	fileRepo repository.UploadRepository,
	folderRepo repository.FolderRepository,
	quotaClient quota.Client,
) *UploadHandler {
	return &UploadHandler{
		fileRepository:   fileRepo,
		folderRepository: folderRepo,
		quotaClient:      quotaClient,
	}
}

type UploadCommand struct {
	Name        string
	Size        uint64
	ContentType string
	ParentID    string
	OwnerID     string
}

func (h *UploadHandler) Handle(
	ctx context.Context,
	cmd UploadCommand,
) (*models.FileMetadata, error) {
	name := strings.TrimSpace(cmd.Name)
	contentType := strings.TrimSpace(cmd.ContentType)
	parentID := strings.TrimSpace(cmd.ParentID)
	ownerID := strings.TrimSpace(cmd.OwnerID)

	if name == "" {
		return nil, ErrInvalidName
	}

	if ownerID == "" {
		return nil, ErrInvalidOwner
	}

	if cmd.Size == 0 || cmd.Size > math.MaxInt64 {
		return nil, ErrInvalidSize
	}

	if contentType == "" {
		return nil, ErrInvalidContentType
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if parentID != "" {
		parent, err := h.folderRepository.GetByID(ctx, parentID)

		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidParent
		}

		if err != nil {
			return nil, fmt.Errorf("get parent folder: %w", err)
		}

		if parent == nil || parent.OwnerID != ownerID {
			return nil, ErrInvalidParent
		}
	}

	existing, err := h.fileRepository.GetByNameAndParent(
		ctx,
		name,
		parentID,
		ownerID,
	)

	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("find existing file: %w", err)
	}

	now := time.Now().UTC().Truncate(time.Millisecond)

	if existing == nil {
		idBytes := make([]byte, 16)

		if _, err := rand.Read(idBytes); err != nil {
			return nil, fmt.Errorf("generate file ID: %w", err)
		}

		id := hex.EncodeToString(idBytes)

		metadata := &models.FileMetadata{
			ID:          id,
			Name:        name,
			Size:        cmd.Size,
			ContentType: contentType,
			StorageKey:  "placeholder/" + id,
			ParentID:    parentID,
			OwnerID:     ownerID,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		return h.save(
			ctx,
			metadata,
			int64(cmd.Size),
			func() error {
				return h.fileRepository.Create(ctx, metadata)
			},
		)
	}

	if existing.Size > math.MaxInt64 {
		return nil, ErrInvalidSize
	}

	delta := int64(cmd.Size) - int64(existing.Size)

	updated := *existing

	updated.Size = cmd.Size
	updated.ContentType = contentType
	updated.StorageKey = "placeholder/" + existing.ID

	if !now.After(existing.UpdatedAt) {
		now = existing.UpdatedAt.Add(time.Millisecond)
	}

	updated.UpdatedAt = now

	return h.save(
		ctx,
		&updated,
		delta,
		func() error {
			return h.fileRepository.UpdateVersion(
				ctx,
				&updated,
				existing.UpdatedAt,
			)
		},
	)
}

func (h *UploadHandler) save(
	ctx context.Context,
	metadata *models.FileMetadata,
	delta int64,
	persist func() error,
) (*models.FileMetadata, error) {
	if delta >= 0 {
		_, err := h.quotaClient.AdjustStorageUsage(
			ctx,
			metadata.OwnerID,
			delta,
		)

		if err != nil {
			return nil, fmt.Errorf("reserve storage: %w", err)
		}

		if err := persist(); err != nil {
			if delta == 0 {
				return nil, err
			}

			rollbackCtx, cancel := context.WithTimeout(
				context.WithoutCancel(ctx),
				5*time.Second,
			)
			defer cancel()

			_, rollbackErr := h.quotaClient.AdjustStorageUsage(
				rollbackCtx,
				metadata.OwnerID,
				-delta,
			)

			if rollbackErr != nil {
				return nil, errors.Join(
					err,
					fmt.Errorf(
						"storage rollback failed: %w",
						rollbackErr,
					),
				)
			}

			return nil, err
		}

		return metadata, nil
	}

	if _, err := h.quotaClient.GetStorage(
		ctx,
		metadata.OwnerID,
	); err != nil {
		return nil, fmt.Errorf("check storage: %w", err)
	}

	if err := persist(); err != nil {
		return nil, err
	}

	releaseCtx, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		5*time.Second,
	)
	defer cancel()

	_, err := h.quotaClient.AdjustStorageUsage(
		releaseCtx,
		metadata.OwnerID,
		delta,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"metadata saved but storage release failed: %w",
			err,
		)
	}

	return metadata, nil
}
