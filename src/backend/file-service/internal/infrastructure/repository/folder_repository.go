package repository

import (
	"context"
	"fmt"

	"file-service/internal/domain/models"
	"file-service/internal/domain/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type folderRepository struct {
	collection *mongo.Collection
}

func NewFolderRepository(db *mongo.Database) repository.FolderRepository {
	return &folderRepository{
		collection: db.Collection("folders"),
	}
}

func (r *folderRepository) Create(
	ctx context.Context,
	folder *models.FolderMetadata,
) error {
	if folder == nil {
		return fmt.Errorf("folder metadata cannot be nil")
	}

	_, err := r.collection.InsertOne(ctx, folder)
	if err != nil {
		return fmt.Errorf("failed to save folder metadata: %w", err)
	}

	return nil
}

func (r *folderRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.FolderMetadata, error) {
	var folder models.FolderMetadata

	err := r.collection.FindOne(
		ctx,
		bson.M{"_id": id},
	).Decode(&folder)

	if err != nil {
		return nil, fmt.Errorf("failed to get folder metadata: %w", err)
	}

	return &folder, nil
}
