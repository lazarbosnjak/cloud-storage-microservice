package mongodb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"user-service/internal/domain"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UserRepository struct {
	collection *mongo.Collection
	logger     *slog.Logger
}

func NewUserRepository(db *mongo.Database, logger *slog.Logger) *UserRepository {
	return &UserRepository{
		collection: db.Collection("user"),
		logger:     logger,
	}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("mongo insert user error: %w", err)
	}

	return nil
}

func (r *UserRepository) GetStorage(
	ctx context.Context,
	userID string,
) (domain.StorageInfo, error) {
	var result struct {
		Quota *int64 `bson:"storageQuota"`
		Used  *int64 `bson:"storageUsed"`
	}

	err := r.collection.FindOne(
		ctx,
		bson.M{"_id": userID},
	).Decode(&result)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.StorageInfo{}, domain.ErrUserNotFound
	}

	if err != nil {
		return domain.StorageInfo{}, fmt.Errorf("get user storage: %w", err)
	}

	if result.Quota == nil || result.Used == nil {
		return domain.StorageInfo{}, domain.ErrStorageNotInitialized
	}

	if *result.Quota <= 0 || *result.Used < 0 {
		return domain.StorageInfo{}, domain.ErrInvalidStorageUsage
	}

	return domain.StorageInfo{
		Quota: *result.Quota,
		Used:  *result.Used,
	}, nil
}

func (r *UserRepository) AdjustStorageUsage(
	ctx context.Context,
	userID string,
	delta int64,
) (domain.StorageInfo, error) {
	newUsage := bson.M{
		"$add": bson.A{"$storageUsed", delta},
	}

	filter := bson.M{
		"_id": userID,
		"storageQuota": bson.M{
			"$type": "long",
			"$gt":   int64(0),
		},
		"storageUsed": bson.M{
			"$type": "long",
			"$gte":  int64(0),
		},
		"$expr": bson.M{
			"$and": bson.A{
				bson.M{
					"$gte": bson.A{newUsage, int64(0)},
				},
				bson.M{
					"$lte": bson.A{newUsage, "$storageQuota"},
				},
			},
		},
	}

	update := bson.M{
		"$inc": bson.M{
			"storageUsed": delta,
		},
	}

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var storage domain.StorageInfo

	err := r.collection.FindOneAndUpdate(
		ctx,
		filter,
		update,
		opts,
	).Decode(&storage)

	if errors.Is(err, mongo.ErrNoDocuments) {
		current, getErr := r.GetStorage(ctx, userID)
		if getErr != nil {
			return domain.StorageInfo{}, getErr
		}

		if delta < -current.Used {
			return domain.StorageInfo{}, domain.ErrInvalidStorageUsage
		}

		if delta > current.Quota-current.Used {
			return domain.StorageInfo{}, domain.ErrInsufficientStorage
		}

		return domain.StorageInfo{}, domain.ErrInvalidStorageUsage
	}

	if err != nil {
		return domain.StorageInfo{}, fmt.Errorf("adjust storage usage: %w", err)
	}

	return storage, nil
}
