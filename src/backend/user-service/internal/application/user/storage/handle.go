package storage

import (
	"context"
	"errors"
	"strings"

	"user-service/internal/domain"
)

var ErrInvalidUserID = errors.New("invalid user ID")

type Repository interface {
	GetStorage(
		context.Context,
		string,
	) (domain.StorageInfo, error)

	AdjustStorageUsage(
		context.Context,
		string,
		int64,
	) (domain.StorageInfo, error)
}

type Handler struct {
	repository Repository
}

func NewHandler(repository Repository) *Handler {
	return &Handler{
		repository: repository,
	}
}

func (h *Handler) GetStorage(
	ctx context.Context,
	userID string,
) (domain.StorageInfo, error) {
	userID = strings.TrimSpace(userID)

	if userID == "" {
		return domain.StorageInfo{}, ErrInvalidUserID
	}

	return h.repository.GetStorage(ctx, userID)
}

func (h *Handler) AdjustStorageUsage(
	ctx context.Context,
	userID string,
	delta int64,
) (domain.StorageInfo, error) {
	userID = strings.TrimSpace(userID)

	if userID == "" {
		return domain.StorageInfo{}, ErrInvalidUserID
	}

	if delta == 0 {
		return h.repository.GetStorage(ctx, userID)
	}

	return h.repository.AdjustStorageUsage(ctx, userID, delta)
}
