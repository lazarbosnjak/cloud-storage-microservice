package repository

import (
	"context"

	"file-service/internal/domain/models"
)

type FolderRepository interface {
	Create(ctx context.Context, folder *models.FolderMetadata) error
	GetByID(ctx context.Context, id string) (*models.FolderMetadata, error)
}
