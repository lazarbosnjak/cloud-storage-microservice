package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"testing"
	"time"

	"file-service/internal/domain/models"
	"file-service/internal/infrastructure/database"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func testID(t *testing.T) string {
	t.Helper()

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}

func setupTestDB(t *testing.T) (*mongo.Database, context.Context) {
	t.Helper()

	if os.Getenv("RUN_MONGO_INTEGRATION") != "1" {
		t.Skip("MongoDB integration tests are disabled")
	}

	if os.Getenv("MONGO_DB") != "files_test_db" {
		t.Fatal("MONGO_DB must be files_test_db")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	client, err := database.ConnectMongoDB(ctx)
	if err != nil {
		t.Fatalf("MongoDB connection failed: %v", err)
	}

	t.Cleanup(func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(
			context.Background(), 10*time.Second,
		)
		defer disconnectCancel()

		if err := client.Disconnect(disconnectCtx); err != nil {
			t.Errorf("MongoDB disconnect failed: %v", err)
		}
	})

	return client.Database("files_test_db"), ctx
}

func TestFileRepositoryIntegration(t *testing.T) {
	db, ctx := setupTestDB(t)
	repo := NewFileRepository(db)

	id := testID(t)
	now := time.Now().UTC().Truncate(time.Millisecond)

	file := &models.FileMetadata{
		ID:        id,
		Name:      "integration-test-file.txt",
		ParentID:  "test-parent",
		OwnerID:   "test-owner",
		CreatedAt: now,
		UpdatedAt: now,
	}

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(
			context.Background(), 10*time.Second,
		)
		defer cancel()

		if _, err := db.Collection("files").DeleteOne(
			cleanupCtx, bson.M{"_id": id},
		); err != nil {
			t.Errorf("File cleanup failed: %v", err)
		}
	})

	if err := repo.Create(ctx, file); err != nil {
		t.Fatalf("File Create failed: %v", err)
	}

	saved, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("File GetByID failed: %v", err)
	}

	if saved.ID != file.ID ||
		saved.Name != file.Name ||
		saved.ParentID != file.ParentID ||
		saved.OwnerID != file.OwnerID ||
		!saved.CreatedAt.Equal(file.CreatedAt) ||
		!saved.UpdatedAt.Equal(file.UpdatedAt) {
		t.Fatalf("File data mismatch: got %+v, want %+v", saved, file)
	}

	t.Log("File successfully saved and retrieved from MongoDB")
}

func TestFolderRepositoryIntegration(t *testing.T) {
	db, ctx := setupTestDB(t)
	repo := NewFolderRepository(db)

	id := testID(t)
	now := time.Now().UTC().Truncate(time.Millisecond)

	folder := &models.FolderMetadata{
		ID:          id,
		Name:        "integration-test-folder",
		Size:        1024,
		ContentType: "application/octet-stream",
		StorageKey:  "test-storage-key",
		ParentID:    "test-parent",
		OwnerID:     "test-owner",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(
			context.Background(), 10*time.Second,
		)
		defer cancel()

		if _, err := db.Collection("folders").DeleteOne(
			cleanupCtx, bson.M{"_id": id},
		); err != nil {
			t.Errorf("Folder cleanup failed: %v", err)
		}
	})

	if err := repo.Create(ctx, folder); err != nil {
		t.Fatalf("Folder Create failed: %v", err)
	}

	saved, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("Folder GetByID failed: %v", err)
	}

	if saved.ID != folder.ID ||
		saved.Name != folder.Name ||
		saved.Size != folder.Size ||
		saved.ContentType != folder.ContentType ||
		saved.StorageKey != folder.StorageKey ||
		saved.ParentID != folder.ParentID ||
		saved.OwnerID != folder.OwnerID ||
		!saved.CreatedAt.Equal(folder.CreatedAt) ||
		!saved.UpdatedAt.Equal(folder.UpdatedAt) {
		t.Fatalf("Folder data mismatch: got %+v, want %+v", saved, folder)
	}

	t.Log("Folder successfully saved and retrieved from MongoDB")
}

func TestFileRepositoryGetByIDNotFound(t *testing.T) {
	db, ctx := setupTestDB(t)
	repo := NewFileRepository(db)

	id := testID(t)

	file, err := repo.GetByID(ctx, id)

	if !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("Expected mongo.ErrNoDocuments, got: %v", err)
	}

	if file != nil {
		t.Fatalf("Expected nil file, got: %+v", file)
	}

	t.Log("Non-existing file correctly returns mongo.ErrNoDocuments")
}

func TestFolderRepositoryGetByIDNotFound(t *testing.T) {
	db, ctx := setupTestDB(t)
	repo := NewFolderRepository(db)

	id := testID(t)

	folder, err := repo.GetByID(ctx, id)

	if !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("Expected mongo.ErrNoDocuments, got: %v", err)
	}

	if folder != nil {
		t.Fatalf("Expected nil folder, got: %+v", folder)
	}

	t.Log("Non-existing folder correctly returns mongo.ErrNoDocuments")
}
