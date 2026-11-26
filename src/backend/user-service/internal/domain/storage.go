package domain

import "errors"

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrInsufficientStorage   = errors.New("insufficient storage")
	ErrInvalidStorageUsage   = errors.New("invalid storage usage")
	ErrStorageNotInitialized = errors.New("storage not initialized")
)
