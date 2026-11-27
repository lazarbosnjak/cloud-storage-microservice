package file

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"file-service/internal/domain/models"
	"file-service/internal/domain/repository"
)

type GetHandler struct {
	repository repository.FileRepository
}

func NewGetHandler(repo repository.FileRepository) *GetHandler {
	return &GetHandler{repository: repo}
}

type GetQuery struct {
	ID      string
	OwnerID string
}

func (h *GetHandler) Handle(
	ctx context.Context,
	query GetQuery,
) (*models.FileMetadata, error) {
	id := strings.TrimSpace(query.ID)
	ownerID := strings.TrimSpace(query.OwnerID)

	if id == "" {
		return nil, ErrInvalidFileID
	}

	if ownerID == "" {
		return nil, ErrInvalidOwner
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result, err := h.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}

		return nil, fmt.Errorf("get file: %w", err)
	}

	if result == nil || result.OwnerID != ownerID {
		return nil, repository.ErrNotFound
	}

	return result, nil
}
