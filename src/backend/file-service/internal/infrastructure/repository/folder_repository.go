package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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
		slog.Warn("cannot save nil folder metadata")
		return fmt.Errorf("folder metadata cannot be nil")
	}

	_, err := r.collection.InsertOne(ctx, folder)
	if err != nil {
		slog.Error("failed to insert folder metadata into MongoDB", "err", err)
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
		slog.Error("failed to retrieve folder metadata from MongoDB", "err", err)
		return nil, fmt.Errorf("failed to get folder metadata: %w", err)
	}

	return &folder, nil
}

func (r *folderRepository) Rename(
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
		slog.Error("failed to rename folder in MongoDB", "err", err)
		return fmt.Errorf("failed to rename folder: %w", err)
	}

	if result.MatchedCount == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *folderRepository) ListByParentID(
	ctx context.Context,
	parentID string,
	ownerID string,
) ([]*models.FolderMetadata, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"ParentId": parentID,
		"ownerId":  ownerID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list folders: %w", err)
	}
	defer cursor.Close(ctx)

	var folders []*models.FolderMetadata
	if err := cursor.All(ctx, &folders); err != nil {
		return nil, fmt.Errorf("failed to decode folders: %w", err)
	}

	return folders, nil
}
