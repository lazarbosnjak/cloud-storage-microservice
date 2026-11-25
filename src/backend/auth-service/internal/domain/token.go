package domain

import "time"

type Token struct {
	PrivateKeyPath     string
	KeyID              string
	Issuer             string
	Audience           string
	AccessTTL          time.Duration
	RefreshIdleTTL     time.Duration
	RefreshAbsoluteTTL time.Duration
	RefreshReuseGrace  time.Duration
	CleanupInterval    time.Duration
}
