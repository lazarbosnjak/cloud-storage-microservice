package models

import "time"

type FileMetadata struct {
	ID          string    `bson:"_id"`
	Name        string    `bson:"name"`
	Size        int64     `bson:"size"`
	ContentType string    `bson:"contentType"`
	StorageKey  string    `bson:"storageKey"`
	FolderID    string    `bson:"folderId"`
	OwnerID     string    `bson:"ownerId"`
	CreatedAt   time.Time `bson:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt"`
}
