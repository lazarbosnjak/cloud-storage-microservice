package models

import "time"

type FileMetadata struct {
	ID        string    `bson:"_id"`
	Name      string    `bson:"name"`
	ParentID  string    `bson:"parentFolderId"`
	OwnerID   string    `bson:"ownerId"`
	CreatedAt time.Time `bson:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt"`
}
