package models

import "time"

type FolderMetadata struct {
	ID             string    `bson:"_id"`
	Name           string    `bson:"name"`
	ParentFolderID string    `bson:"parentFolderId"`
	OwnerID        string    `bson:"ownerId"`
	CreatedAt      time.Time `bson:"createdAt"`
	UpdatedAt      time.Time `bson:"updatedAt"`
}
