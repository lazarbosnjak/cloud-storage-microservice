package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"file-service/internal/domain/models"
	"file-service/internal/domain/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type fileRepository struct {
	collection *mongo.Collection
}

func NewFileRepository(
	db *mongo.Database,
) repository.UploadRepository {
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

	if mongo.IsDuplicateKeyError(err) {
		return repository.ErrFileAlreadyExists
	}

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
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
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
	if mongo.IsDuplicateKeyError(err) {
		return repository.ErrFileAlreadyExists
	}

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

func (r *fileRepository) GetByNameAndParent(
	ctx context.Context,
	name string,
	parentID string,
	ownerID string,
) (*models.FileMetadata, error) {
	var file models.FileMetadata

	err := r.collection.FindOne(
		ctx,
		bson.M{
			"name":           name,
			"parentFolderId": parentID,
			"ownerId":        ownerID,
		},
	).Decode(&file)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, repository.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf(
			"failed to find file by name: %w",
			err,
		)
	}

	return &file, nil
}

func (r *fileRepository) UpdateVersion(
	ctx context.Context,
	file *models.FileMetadata,
	expectedUpdatedAt time.Time,
) error {
	if file == nil {
		return fmt.Errorf("file metadata cannot be nil")
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{
			"_id":            file.ID,
			"ownerId":        file.OwnerID,
			"parentFolderId": file.ParentID,
			"name":           file.Name,
			"updatedAt":      expectedUpdatedAt,
		},
		bson.M{
			"$set": bson.M{
				"size":        file.Size,
				"contentType": file.ContentType,
				"storageKey":  file.StorageKey,
				"updatedAt":   file.UpdatedAt,
			},
		},
	)

	if err != nil {
		return fmt.Errorf(
			"failed to update file version: %w",
			err,
		)
	}

	if result.MatchedCount == 0 {
		return repository.ErrVersionConflict
	}

	return nil
}

func EnsureFileIndexes(
	ctx context.Context,
	db *mongo.Database,
) error {
	_, err := db.Collection("files").Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "ownerId", Value: 1},
				{Key: "parentFolderId", Value: 1},
				{Key: "name", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("unique_file_owner_parent_name"),
		},
	)

	if err != nil {
		return fmt.Errorf("create file indexes: %w", err)
	}

	return nil
}
