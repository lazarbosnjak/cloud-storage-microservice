package model

import (
	"github.com/gocql/gocql"
)

type NotificationEventType string

const (
	ResourceShared          NotificationEventType = "RESOURCE_SHARED"
	FileAddedToSharedFolder NotificationEventType = "FILE_ADDED_TO_SHARED_FOLDER"
	NewSharedFileVersion    NotificationEventType = "NEW_SHARED_FILE_VERSION"
)

type NotificationByUserID struct {
	UserID               gocql.UUID            `json:"userId"`
	NotificationID       gocql.UUID            `json:"notificationId"`
	EventType            NotificationEventType `json:"eventType"`
	PayloadSchemaVersion int                   `json:"payloadSchemaVersion"`
	Payload              map[string]any        `json:"payload"`
	ActorID              gocql.UUID            `json:"actorId"`
}
