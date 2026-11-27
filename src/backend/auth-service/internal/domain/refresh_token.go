package domain

import "time"

type RefreshToken struct {
	ID              string
	UserID          string
	FamilyID        string
	TokenHash       []byte
	CreatedAt       time.Time
	ExpiresAt       time.Time
	FamilyExpiresAt time.Time
	UsedAt          *time.Time
	RevokedAt       *time.Time

	UserRole  Role
	IsEnabled bool
}
