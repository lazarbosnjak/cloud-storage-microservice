package database

import (
	"context"
	"fmt"
	"net/url"

	"file-service/internal/startup/config"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func ConnectMongoDB(ctx context.Context, cfg *config.Config) (*mongo.Client, error) {
	mongoURI := fmt.Sprintf(
		"mongodb://%s:%s@%s:%s/?authSource=admin",
		url.QueryEscape(cfg.MongoUsername),
		url.QueryEscape(cfg.MongoPassword),
		cfg.MongoHost,
		cfg.MongoPort,
	)

	client, err := mongo.Connect(
		options.Client().ApplyURI(mongoURI),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create MongoDB client: %w",
			err,
		)
	}

	pingCtx, cancel := context.WithTimeout(
		ctx,
		cfg.MongoPingTimeout,
	)
	defer cancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf(
			"failed to ping MongoDB: %w",
			err,
		)
	}

	return client, nil
}
