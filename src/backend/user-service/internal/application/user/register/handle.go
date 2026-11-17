package register

import (
	"context"

	"user-service/internal/domain"
)

type RegisterUser interface {
	Create(context.Context, domain.User) error
}

type Query struct {
	UserID    string
	Email     string
	FirstName string
	LastName  string
}

type Handler struct {
	repository RegisterUser
}

func NewHandler(repository RegisterUser) *Handler {
	return &Handler{
		repository: repository,
	}
}

func (h *Handler) Register(ctx context.Context, query Query) error {
	user := domain.User{
		MongoID:   "",
		ID:        query.UserID,
		Email:     query.Email,
		FirstName: query.FirstName,
		LastName:  query.LastName,
	}

	return h.repository.Create(ctx, user)
}
