package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"file-service/internal/domain/models"
	domainrepo "file-service/internal/domain/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func cleanupMetadataDocument(
	t *testing.T,
	collection *mongo.Collection,
	id string,
) {
	t.Helper()

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if _, err := collection.DeleteOne(
			ctx,
			bson.M{"_id": id},
		); err != nil {
			t.Errorf("cleanup failed for %s: %v", id, err)
		}
	})
}

func TestFileRepositoryRenameIntegration(t *testing.T) {
	db, ctx := setupTestDB(t)
	repo := NewFileRepository(db)

	id := testID(t)
	createdAt := time.Now().UTC().Truncate(time.Millisecond)
	updatedAt := createdAt.Add(time.Minute)

	cleanupMetadataDocument(t, db.Collection("files"), id)

	err := repo.Create(ctx, &models.FileMetadata{
		ID:        id,
		Name:      "old.txt",
		OwnerID:   "test-owner",
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = repo.Rename(
		ctx,
		id,
		"test-owner",
		"new.txt",
		updatedAt,
	)
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	saved, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	if saved.Name != "new.txt" {
		t.Fatalf("expected new.txt, got %s", saved.Name)
	}

	if !saved.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("UpdatedAt was not updated")
	}

	if !saved.CreatedAt.Equal(createdAt) {
		t.Fatal("CreatedAt should not change")
	}

	err = repo.Rename(
		ctx,
		id,
		"another-owner",
		"forbidden.txt",
		updatedAt,
	)

	if !errors.Is(err, domainrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	saved, err = repo.GetByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	if saved.Name != "new.txt" {
		t.Fatal("unauthorized rename changed the file")
	}
}

func TestFolderRepositoryRenameIntegration(t *testing.T) {
	db, ctx := setupTestDB(t)
	repo := NewFolderRepository(db)

	id := testID(t)
	createdAt := time.Now().UTC().Truncate(time.Millisecond)
	updatedAt := createdAt.Add(time.Minute)

	cleanupMetadataDocument(t, db.Collection("folders"), id)

	err := repo.Create(ctx, &models.FolderMetadata{
		ID:        id,
		Name:      "OldFolder",
		OwnerID:   "test-owner",
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = repo.Rename(
		ctx,
		id,
		"test-owner",
		"NewFolder",
		updatedAt,
	)
	if err != nil {
		t.Fatal(err)
	}

	saved, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	if saved.Name != "NewFolder" {
		t.Fatalf("unexpected folder name: %s", saved.Name)
	}

	if !saved.UpdatedAt.Equal(updatedAt) {
		t.Fatal("UpdatedAt was not updated")
	}

	err = repo.Rename(
		ctx,
		id,
		"another-owner",
		"Forbidden",
		updatedAt,
	)

	if !errors.Is(err, domainrepo.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	saved, err = repo.GetByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	if saved.Name != "NewFolder" {
		t.Fatal("unauthorized rename changed the folder")
	}
}

func TestFileRepositoryListByParentIDIntegration(t *testing.T) {
	db, ctx := setupTestDB(t)
	repo := NewFileRepository(db)

	parentID := testID(t)
	ownerID := testID(t)

	firstID := testID(t)
	secondID := testID(t)
	foreignID := testID(t)

	for _, id := range []string{
		firstID, secondID, foreignID,
	} {
		cleanupMetadataDocument(t, db.Collection("files"), id)
	}

	files := []*models.FileMetadata{
		{
			ID:       firstID,
			Name:     "first.txt",
			ParentID: parentID,
			OwnerID:  ownerID,
		},
		{
			ID:       secondID,
			Name:     "second.txt",
			ParentID: parentID,
			OwnerID:  ownerID,
		},
		{
			ID:       foreignID,
			Name:     "foreign.txt",
			ParentID: parentID,
			OwnerID:  "another-owner",
		},
	}

	for _, value := range files {
		if err := repo.Create(ctx, value); err != nil {
			t.Fatal(err)
		}
	}

	result, err := repo.ListByParentID(
		ctx,
		parentID,
		ownerID,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 files, got %d", len(result))
	}

	found := make(map[string]bool)
	for _, value := range result {
		if value.OwnerID != ownerID {
			t.Fatal("another owner's file was returned")
		}
		found[value.ID] = true
	}

	if !found[firstID] || !found[secondID] {
		t.Fatal("expected files were not returned")
	}
}

func TestFolderRepositoryListByParentIDIntegration(t *testing.T) {
	db, ctx := setupTestDB(t)
	repo := NewFolderRepository(db)

	parentID := testID(t)
	ownerID := testID(t)

	firstID := testID(t)
	secondID := testID(t)
	foreignID := testID(t)

	for _, id := range []string{
		firstID, secondID, foreignID,
	} {
		cleanupMetadataDocument(t, db.Collection("folders"), id)
	}

	folders := []*models.FolderMetadata{
		{
			ID:       firstID,
			Name:     "First",
			ParentID: parentID,
			OwnerID:  ownerID,
		},
		{
			ID:       secondID,
			Name:     "Second",
			ParentID: parentID,
			OwnerID:  ownerID,
		},
		{
			ID:       foreignID,
			Name:     "Foreign",
			ParentID: parentID,
			OwnerID:  "another-owner",
		},
	}

	for _, value := range folders {
		if err := repo.Create(ctx, value); err != nil {
			t.Fatal(err)
		}
	}

	result, err := repo.ListByParentID(
		ctx,
		parentID,
		ownerID,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 folders, got %d", len(result))
	}

	found := make(map[string]bool)
	for _, value := range result {
		if value.OwnerID != ownerID {
			t.Fatal("another owner's folder was returned")
		}
		found[value.ID] = true
	}

	if !found[firstID] || !found[secondID] {
		t.Fatal("expected folders were not returned")
	}
}
