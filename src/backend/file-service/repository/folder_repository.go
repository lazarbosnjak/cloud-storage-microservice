package repository

import (
	"context"
	"fmt"

	"file-service/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type FolderRepository struct {
	collection *mongo.Collection
}

func NewFolderRepository(db *mongo.Database) *FolderRepository {
	return &FolderRepository{
		collection: db.Collection("folders"),
	}
}

// Create
func (r *FolderRepository) Create(
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

// GetByID
func (r *FolderRepository) GetByID(
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
