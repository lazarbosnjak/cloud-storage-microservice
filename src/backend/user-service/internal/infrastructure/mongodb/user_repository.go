package mongodb

import (
	"context"
	"fmt"
	"log/slog"

	"user-service/internal/domain"

	"go.mongodb.org/mongo-driver/v2/mongo"
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
