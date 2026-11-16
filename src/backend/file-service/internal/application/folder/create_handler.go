package folder

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
	ErrInvalidName   = errors.New("folder name cannot be empty")
	ErrInvalidOwner  = errors.New("folder owner cannot be empty")
	ErrInvalidParent = errors.New("parent folder not found or inaccessible")
)

type CreateHandler struct {
	repository repository.FolderRepository
}

func NewCreateHandler(repo repository.FolderRepository) *CreateHandler {
	return &CreateHandler{repository: repo}
}

type CreateCommand struct {
	Name     string
	ParentID string
	OwnerID  string
}

func (h *CreateHandler) Handle(
	ctx context.Context,
	cmd CreateCommand,
) (*models.FolderMetadata, error) {
	name := strings.TrimSpace(cmd.Name)
	ownerID := strings.TrimSpace(cmd.OwnerID)
	parentID := strings.TrimSpace(cmd.ParentID)

	if name == "" {
		return nil, ErrInvalidName
	}

	if ownerID == "" {
		return nil, ErrInvalidOwner
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if parentID != "" {
		parent, err := h.repository.GetByID(ctx, parentID)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidParent, err)
		}

		if parent == nil || parent.OwnerID != ownerID {
			return nil, ErrInvalidParent
		}
	}

	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, fmt.Errorf("generate folder ID: %w", err)
	}

	now := time.Now().UTC()

	folder := &models.FolderMetadata{
		ID:        hex.EncodeToString(idBytes),
		Name:      name,
		ParentID:  parentID,
		OwnerID:   ownerID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.repository.Create(ctx, folder); err != nil {
		return nil, fmt.Errorf("create folder: %w", err)
	}

	return folder, nil
}
