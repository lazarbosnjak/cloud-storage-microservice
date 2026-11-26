package model

import (
	"github.com/gocql/gocql"
)

type NotificationEventType string

const (
	TypeFileShared              NotificationEventType = "FILE_SHARED"
	TypeFolderShared            NotificationEventType = "FOLDER_SHARED"
	TypeFileAddedToSharedFolder NotificationEventType = "FILE_ADDED_TO_SHARED_FOLDER"
	TypeSharedFileNewVersion    NotificationEventType = "SHARED_FILE_NEW_VERSION"
)

type Payload interface {
	EventType() NotificationEventType
	SchemaVersion() int
}

type NotificationByUserID struct {
	UserID         gocql.UUID
	NotificationID gocql.UUID
	EventType      NotificationEventType
	Payload        Payload
}

// PAYLOAD IMPLEMENTATIONS

type FolderShared struct {
	FolderID   string `json:"folderId"`
	FolderName string `json:"folderName"`
	ActorID    string `json:"actorId"`
	ActorName  string `json:"actorName"`
}

func (FolderShared) SchemaVersion() int               { return 1 }
func (FolderShared) EventType() NotificationEventType { return TypeFolderShared }

type FileAddedToSharedFolder struct {
	FileID           string `json:"fileId"`
	FileName         string `json:"fileName"`
	SharedFolderID   string `json:"sharedFolderId"`
	SharedFolderName string `json:"sharedFolderName"`
	ActorID          string `json:"actorId"`
	ActorName        string `json:"actorName"`
}

func (FileAddedToSharedFolder) SchemaVersion() int { return 1 }
func (FileAddedToSharedFolder) EventType() NotificationEventType {
	return TypeFileAddedToSharedFolder
}

type FileShared struct {
	FileID    string `json:"fileId"`
	FileName  string `json:"fileName"`
	ActorID   string `json:"actorId"`
	ActorName string `json:"actorName"`
}

func (FileShared) SchemaVersion() int               { return 1 }
func (FileShared) EventType() NotificationEventType { return TypeFileShared }

type SharedFileNewVersion struct {
	FileID    string `json:"fileId"`
	FileName  string `json:"fileName"`
	ActorID   string `json:"actorId"`
	ActorName string `json:"actorName"`
}

func (SharedFileNewVersion) SchemaVersion() int { return 1 }
func (SharedFileNewVersion) EventType() NotificationEventType {
	return TypeSharedFileNewVersion
}
