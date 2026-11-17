package repository

import (
	"context"
	"fmt"
	"time"

	"file-service/internal/domain/models"
	"file-service/internal/domain/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
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

func (r *fileRepository) Rename(
	ctx context.Context,
	id string,
	ownerID string,
	name string,
	updatedAt time.Time,
) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{
			"_id":     id,
			"ownerId": ownerID,
		},
		bson.M{
			"$set": bson.M{
				"name":      name,
				"updatedAt": updatedAt,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	if result.MatchedCount == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *fileRepository) ListByParentID(
	ctx context.Context,
	parentID string,
	ownerID string,
) ([]*models.FileMetadata, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"parentFolderId": parentID,
		"ownerId":        ownerID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	defer cursor.Close(ctx)

	var files []*models.FileMetadata
	if err := cursor.All(ctx, &files); err != nil {
		return nil, fmt.Errorf("failed to decode files: %w", err)
	}

	return files, nil
}
