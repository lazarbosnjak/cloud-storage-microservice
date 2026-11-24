package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

func getString(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		slog.Warn("reading env", key, "not found")
		os.Exit(1)
	}

	return val
}

func getUInt(key string) uint32 {
	return uint32(getInt(key))
}

func getInt(key string) int32 {
	val := getString(key)
	i32, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		slog.Error("parsing to int", "value", val)
		os.Exit(1)
	}
	return int32(i32)
}

func getDuration(key string) time.Duration {
	raw := getString(key)

	d, err := time.ParseDuration(raw)
	if err != nil {
		slog.Error("parsing to duration", "key", key, "value", raw)
		os.Exit(1)
	}

	return d
}
