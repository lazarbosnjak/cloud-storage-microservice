package config

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

type Lockout struct {
	Treshold     int
	BaseDuration time.Duration
	MaxDuration  time.Duration
	ResetWindow  time.Duration
}

func loadToken(l *loader) Token {
	t := Token{
		PrivateKeyPath:     l.required("JWT_PRIVATE_KEY_PATH"),
		KeyID:              l.str("JWT_KEY_ID", "dev-1"),
		Issuer:             l.str("JWT_ISSUER", "auth-service"),
		Audience:           l.str("JWT_AUDIENCE", "pouch"),
		AccessTTL:          l.duration("ACCESS_TOKEN_TTL", 10*time.Minute),
		RefreshIdleTTL:     l.duration("REFRESH_TOKEN_IDLE_TTL", 168*time.Hour),
		RefreshAbsoluteTTL: l.duration("REFRESH_TOKEN_ABSOLUTE_TTL", 720*time.Hour),
		RefreshReuseGrace:  l.duration("REFRESH_REUSE_GRACE", 10*time.Second),
		CleanupInterval:    l.duration("REFRESH_CLEANUP_INTERVAL", time.Hour),
	}

	if t.RefreshIdleTTL > t.RefreshAbsoluteTTL {
		l.fail("REFRESH_TOKEN_IDLE_TTL must not exceed REFRESH_TOKEN_ABSOLUTE_TTL")
	}

	return t
}

func loadLockout(l *loader) Lockout {
	lo := Lockout{
		Treshold:     l.integer("LOCKOUT_TRESHOLD", 5),
		BaseDuration: l.duration("LOCKOUT_BASE_DURATION", time.Minute),
		MaxDuration:  l.duration("LOCKOUT_MAX_DURATION", 15*time.Minute),
		ResetWindow:  l.duration("LOCKOUT_RESET_WINDOW", 15*time.Minute),
	}
	if lo.Treshold < 1 {
		l.fail("LOCKOUT_TRESHOLD must be at least 1")
	}
	if lo.BaseDuration > lo.MaxDuration {
		l.fail("LOCKOUT_BASE_DURATION must not exceed LOCKOUT_MAX_DURATION")
	}

	return lo
}
