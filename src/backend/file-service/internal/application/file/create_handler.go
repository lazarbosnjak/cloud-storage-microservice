package file

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"file-service/internal/domain/models"
	"file-service/internal/domain/repository"
)

var (
	ErrInvalidSize        = errors.New("file size must be greater than zero")
	ErrInvalidContentType = errors.New("file content type cannot be empty")
	ErrInvalidStorageKey  = errors.New("file storage key cannot be empty")
	ErrInvalidParent      = errors.New("parent folder not found or inaccessible")
)

type CreateHandler struct {
	fileRepository   repository.FileRepository
	folderRepository repository.FolderRepository
}

func NewCreateHandler(
	fileRepo repository.FileRepository,
	folderRepo repository.FolderRepository,
) *CreateHandler {
	return &CreateHandler{
		fileRepository:   fileRepo,
		folderRepository: folderRepo,
	}
}

type CreateCommand struct {
	Name        string
	Size        uint64
	ContentType string
	StorageKey  string
	ParentID    string
	OwnerID     string
}

func (h *CreateHandler) Handle(
	ctx context.Context,
	cmd CreateCommand,
) (*models.FileMetadata, error) {
	name := strings.TrimSpace(cmd.Name)
	contentType := strings.TrimSpace(cmd.ContentType)
	storageKey := strings.TrimSpace(cmd.StorageKey)
	parentID := strings.TrimSpace(cmd.ParentID)
	ownerID := strings.TrimSpace(cmd.OwnerID)

	if name == "" {
		return nil, ErrInvalidName
	}

	if ownerID == "" {
		return nil, ErrInvalidOwner
	}

	if cmd.Size == 0 {
		return nil, ErrInvalidSize
	}

	if contentType == "" {
		return nil, ErrInvalidContentType
	}

	if storageKey == "" {
		return nil, ErrInvalidStorageKey
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if parentID != "" {
		parent, err := h.folderRepository.GetByID(ctx, parentID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, ErrInvalidParent
			}
			return nil, fmt.Errorf("get parent folder: %w", err)
		}

		if parent == nil || parent.OwnerID != ownerID {
			return nil, ErrInvalidParent
		}
	}

	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, fmt.Errorf("generate file ID: %w", err)
	}

	now := time.Now().UTC()

	metadata := &models.FileMetadata{
		ID:          hex.EncodeToString(idBytes),
		Name:        name,
		Size:        cmd.Size,
		ContentType: contentType,
		StorageKey:  storageKey,
		ParentID:    parentID,
		OwnerID:     ownerID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.fileRepository.Create(ctx, metadata); err != nil {
		return nil, fmt.Errorf("create file metadata: %w", err)
	}

	return metadata, nil
}
