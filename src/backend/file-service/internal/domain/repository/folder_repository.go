package repository

import (
	"context"
	"time"

	"file-service/internal/domain/models"
)

type FolderRepository interface {
	Create(ctx context.Context, folder *models.FolderMetadata) error
	GetByID(ctx context.Context, id string) (*models.FolderMetadata, error)
	Rename(ctx context.Context, id string, ownerID string, name string, updatedAt time.Time) error
	ListByParentID(ctx context.Context, parentID string, ownerID string) ([]*models.FolderMetadata, error)
}
