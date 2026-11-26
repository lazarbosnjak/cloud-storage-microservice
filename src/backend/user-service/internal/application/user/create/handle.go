package create

import (
	"context"

	"user-service/internal/domain"
)

type RegisterUser interface {
	Create(context.Context, domain.User) error
}

type Query struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
}

type Handler struct {
	repository          RegisterUser
	defaultStorageQuota int64
}

func NewHandler(repository RegisterUser, defaultStorageQuota int64) *Handler {
	return &Handler{
		repository:          repository,
		defaultStorageQuota: defaultStorageQuota,
	}
}

func (h *Handler) Register(ctx context.Context, query Query) error {
	user := domain.User{
		ID:           query.ID,
		Email:        query.Email,
		FirstName:    query.FirstName,
		LastName:     query.LastName,
		StorageQuota: h.defaultStorageQuota,
		StorageUsed:  0,
	}

	return h.repository.Create(ctx, user)
}
