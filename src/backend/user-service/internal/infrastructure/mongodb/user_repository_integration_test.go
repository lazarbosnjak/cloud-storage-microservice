package mongodb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"user-service/internal/domain"
	"user-service/internal/startup/config"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func newIntegrationRepository(t *testing.T) *UserRepository {
	t.Helper()

	if os.Getenv("RUN_MONGO_INTEGRATION") != "1" {
		t.Skip("MongoDB integration tests disabled")
	}

	logger := slog.Default()
	cfg := config.NewConfig(logger)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	client, err := NewClient(ctx, cfg)
	if err != nil {
		t.Fatalf("connect to MongoDB: %v", err)
	}

	collectionName := fmt.Sprintf(
		"user_integration_test_%d",
		time.Now().UnixNano(),
	)

	collection := client.
		Database(cfg.Mongo.Database).
		Collection(collectionName)

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cleanupCancel()

		if err := collection.Drop(cleanupCtx); err != nil {
			t.Errorf("drop test collection: %v", err)
		}

		if err := client.Disconnect(cleanupCtx); err != nil {
			t.Errorf("disconnect MongoDB: %v", err)
		}
	})

	return &UserRepository{
		collection: collection,
		logger:     logger,
	}
}

func createTestUser(
	t *testing.T,
	repository *UserRepository,
	quota int64,
	used int64,
) {
	t.Helper()

	err := repository.Create(context.Background(), domain.User{
		ID:           "test-user",
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		StorageQuota: quota,
		StorageUsed:  used,
	})

	if err != nil {
		t.Fatalf("create test user: %v", err)
	}
}

func TestMongoGetStorage(t *testing.T) {
	repository := newIntegrationRepository(t)
	createTestUser(t, repository, 1000, 200)

	result, err := repository.GetStorage(
		context.Background(),
		"test-user",
	)

	if err != nil {
		t.Fatal(err)
	}

	if result.Quota != 1000 || result.Used != 200 {
		t.Fatalf("unexpected storage: %+v", result)
	}
}

func TestMongoAdjustStorageUsage(t *testing.T) {
	repository := newIntegrationRepository(t)
	createTestUser(t, repository, 1000, 200)

	ctx := context.Background()

	result, err := repository.AdjustStorageUsage(
		ctx, "test-user", 300,
	)

	if err != nil || result.Used != 500 {
		t.Fatalf("increase failed: %+v, %v", result, err)
	}

	result, err = repository.AdjustStorageUsage(
		ctx, "test-user", -100,
	)

	if err != nil || result.Used != 400 {
		t.Fatalf("decrease failed: %+v, %v", result, err)
	}
}

func TestMongoStorageQuotaExceeded(t *testing.T) {
	repository := newIntegrationRepository(t)
	createTestUser(t, repository, 1000, 800)

	ctx := context.Background()

	_, err := repository.AdjustStorageUsage(
		ctx, "test-user", 300,
	)

	if !errors.Is(err, domain.ErrInsufficientStorage) {
		t.Fatalf("expected insufficient storage, got %v", err)
	}

	result, err := repository.GetStorage(ctx, "test-user")

	if err != nil || result.Used != 800 {
		t.Fatalf("usage changed unexpectedly: %+v, %v", result, err)
	}
}

func TestMongoNegativeStorageUsage(t *testing.T) {
	repository := newIntegrationRepository(t)
	createTestUser(t, repository, 1000, 200)

	ctx := context.Background()

	_, err := repository.AdjustStorageUsage(
		ctx, "test-user", -300,
	)

	if !errors.Is(err, domain.ErrInvalidStorageUsage) {
		t.Fatalf("expected invalid usage, got %v", err)
	}

	result, err := repository.GetStorage(ctx, "test-user")

	if err != nil || result.Used != 200 {
		t.Fatalf("usage changed unexpectedly: %+v, %v", result, err)
	}
}

func TestMongoStorageExactQuota(t *testing.T) {
	repository := newIntegrationRepository(t)
	createTestUser(t, repository, 1000, 800)

	result, err := repository.AdjustStorageUsage(
		context.Background(), "test-user", 200,
	)

	if err != nil || result.Used != 1000 {
		t.Fatalf("exact quota failed: %+v, %v", result, err)
	}
}

func TestMongoStorageUserNotFound(t *testing.T) {
	repository := newIntegrationRepository(t)

	ctx := context.Background()

	_, err := repository.GetStorage(ctx, "missing-user")

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected user not found, got %v", err)
	}

	_, err = repository.AdjustStorageUsage(
		ctx, "missing-user", 100,
	)

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected user not found, got %v", err)
	}
}

func TestMongoStorageNotInitialized(t *testing.T) {
	repository := newIntegrationRepository(t)

	ctx := context.Background()

	_, err := repository.collection.InsertOne(ctx, bson.M{
		"_id":   "legacy-user",
		"email": "legacy@example.com",
	})

	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.GetStorage(ctx, "legacy-user")

	if !errors.Is(err, domain.ErrStorageNotInitialized) {
		t.Fatalf("expected storage not initialized, got %v", err)
	}

	_, err = repository.AdjustStorageUsage(
		ctx, "legacy-user", 100,
	)

	if !errors.Is(err, domain.ErrStorageNotInitialized) {
		t.Fatalf("expected storage not initialized, got %v", err)
	}
}

func TestMongoConcurrentStorageAdjustments(t *testing.T) {
	repository := newIntegrationRepository(t)
	createTestUser(t, repository, 1000, 800)

	ctx := context.Background()

	var wg sync.WaitGroup

	results := make(chan error, 2)
	start := make(chan struct{})

	for i := 0; i < 2; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			<-start

			_, err := repository.AdjustStorageUsage(
				ctx, "test-user", 150,
			)

			results <- err
		}()
	}

	close(start)

	wg.Wait()
	close(results)

	successful := 0
	rejected := 0

	for err := range results {
		switch {
		case err == nil:
			successful++

		case errors.Is(err, domain.ErrInsufficientStorage):
			rejected++

		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if successful != 1 || rejected != 1 {
		t.Fatalf(
			"expected 1 success and 1 rejection, got %d and %d",
			successful,
			rejected,
		)
	}

	result, err := repository.GetStorage(ctx, "test-user")

	if err != nil || result.Used != 950 {
		t.Fatalf("unexpected final usage: %+v, %v", result, err)
	}
}
