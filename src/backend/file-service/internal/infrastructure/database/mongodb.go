package database

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func ConnectMongoDB(ctx context.Context) (*mongo.Client, error) {
	host := os.Getenv("MONGO_HOST")
	port := os.Getenv("MONGO_PORT")
	username := os.Getenv("MONGO_INITDB_ROOT_USERNAME")
	password := os.Getenv("MONGO_INITDB_ROOT_PASSWORD")

	if host == "" || port == "" || username == "" || password == "" {
		return nil, fmt.Errorf("missing MongoDB environment variables")
	}

	mongoURI := fmt.Sprintf(
		"mongodb://%s:%s@%s:%s/?authSource=admin",
		url.QueryEscape(username),
		url.QueryEscape(password),
		host,
		port,
	)

	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client, nil
}
