package quota

import (
	"context"
	"errors"
)

var (
	ErrInsufficientStorage = errors.New("insufficient storage")
	ErrUserNotFound        = errors.New("user not found")
	ErrStorageUnavailable  = errors.New("user storage unavailable")
	ErrInvalidStorageUsage = errors.New("invalid storage usage")
)

type StorageInfo struct {
	Quota     int64
	Used      int64
	Available int64
}

type Client interface {
	GetStorage(
		ctx context.Context,
		userID string,
	) (StorageInfo, error)

	AdjustStorageUsage(
		ctx context.Context,
		userID string,
		delta int64,
	) (StorageInfo, error)
}
