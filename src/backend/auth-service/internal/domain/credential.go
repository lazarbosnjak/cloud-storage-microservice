package domain

import "time"

type Role int16

const (
	RoleAdmin Role = 1
	RoleUser  Role = 2
)

func (r Role) String() string {
	switch r {
	case RoleAdmin:
		return "admin"
	case RoleUser:
		return "user"
	default:
		return "unknown"
	}
}

type Credential struct {
	ID                string
	Email             string
	PasswordHash      string
	UserRole          Role
	PasswordChangedAt time.Time
	FailedAttempts    uint8
	LockedUntil       *time.Time
	LastFailedAt      *time.Time
	IsLocked          bool
	LockedReason      string
	CreatedAt         time.Time
	IsEnabled         bool
}
