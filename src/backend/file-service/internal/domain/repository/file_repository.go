package repository

import (
	"context"

	"file-service/internal/domain/models"
)

type FileRepository interface {
	Create(ctx context.Context, file *models.FileMetadata) error
	GetByID(ctx context.Context, id string) (*models.FileMetadata, error)
}
