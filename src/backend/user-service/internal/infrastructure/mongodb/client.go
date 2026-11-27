package mongodb

import (
	"context"
	"fmt"
	"net/url"

	"user-service/internal/startup/config"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewClient(ctx context.Context, cfg *config.Config) (*mongo.Client, error) {
	var mongoURI string
	if cfg.Mongo.InitDBRootUsername != "" {
		mongoURI = fmt.Sprintf(
			"mongodb://%s:%s@%s:%s/?authSource=admin",
			url.QueryEscape(cfg.Mongo.InitDBRootUsername),
			url.QueryEscape(cfg.Mongo.InitDBRootPassword),
			cfg.Mongo.Host,
			cfg.Mongo.Port,
		)
	} else {
		mongoURI = fmt.Sprintf(
			"mongodb://%s:%s",
			cfg.Mongo.Host,
			cfg.Mongo.Port,
		)
	}

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
		cfg.Mongo.PingTimeout,
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
