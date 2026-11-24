package domain

import "time"

type VerificationTokenReason int

const (
	Registration VerificationTokenReason = iota
	MagicLink
)

type VerificationToken struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
	UsedAt    *time.Time
	Reason    VerificationTokenReason
}
