package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type loader struct {
	problems []string
}

func (l *loader) fail(format string, args ...any) {
	l.problems = append(l.problems, fmt.Sprintf(format, args...))
}

func (l *loader) required(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		l.fail("%s is required", key)
	}
	return v
}

func (l *loader) str(key string, def string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	return v
}

func (l *loader) duration(key string, def time.Duration) time.Duration {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		l.fail("%s: invalid duration %q (use values like 10m or 168h)", key, raw)
		return def
	}
	if d <= 0 {
		l.fail("%s must be positive", key)
	}
	return d
}

func (l *loader) integer(key string, def int) int {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		l.fail("%s: invalid integer %q", key, raw)
		return def
	}
	return n
}

func (l *loader) err() error {
	if len(l.problems) == 0 {
		return nil
	}
	return fmt.Errorf("invalid configuration: %s", strings.Join(l.problems, "; "))
}
