package domain

import "time"

type OtpCode struct {
	ID        string
	UserID    string
	CodeHash  []byte
	Attempts  int
	CreatedAt time.Time
	ExpiresAt time.Time
	UsedAt    *time.Time
}
