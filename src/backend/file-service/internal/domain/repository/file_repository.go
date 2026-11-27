package repository

import (
	"context"
	"time"

	"file-service/internal/domain/models"
)

type FileRepository interface {
	Create(
		ctx context.Context,
		file *models.FileMetadata,
	) error

	GetByID(
		ctx context.Context,
		id string) (*models.FileMetadata, error)

	Rename(
		ctx context.Context,
		id string,
		ownerID string,
		name string,
		updatedAt time.Time,
	) error

	ListByParentID(
		ctx context.Context,
		parentID string,
		ownerID string) ([]*models.FileMetadata, error)
}

type UploadRepository interface {
	FileRepository

	GetByNameAndParent(
		ctx context.Context,
		name string,
		parentID string,
		ownerID string,
	) (*models.FileMetadata, error)

	UpdateVersion(
		ctx context.Context,
		file *models.FileMetadata,
		expectedUpdatedAt time.Time,
	) error
}
