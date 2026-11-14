package models

import "time"

type FolderMetadata struct {
	ID          string    `bson:"_id"`
	Name        string    `bson:"name"`
	Size        uint64    `bson:"size"`
	ContentType string    `bson:"contentType"`
	StorageKey  string    `bson:"storageKey"`
	ParentID    string    `bson:"ParentId"`
	OwnerID     string    `bson:"ownerId"`
	CreatedAt   time.Time `bson:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt"`
}
