package repository

import (
	"context"
	"fmt"

	"file-service/internal/domain/models"
	"file-service/internal/domain/repository"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type fileRepository struct {
	collection *mongo.Collection
}

func NewFileRepository(db *mongo.Database) repository.FileRepository {
	return &fileRepository{
		collection: db.Collection("files"),
	}
}

func (r *fileRepository) Create(
	ctx context.Context,
	file *models.FileMetadata,
) error {
	if file == nil {
		return fmt.Errorf("file metadata cannot be nil")
	}

	_, err := r.collection.InsertOne(ctx, file)
	if err != nil {
		return fmt.Errorf("failed to save file metadata: %w", err)
	}

	return nil
}

func (r *fileRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.FileMetadata, error) {
	var file models.FileMetadata

	err := r.collection.FindOne(
		ctx,
		map[string]interface{}{"_id": id},
	).Decode(&file)

	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	return &file, nil
}
