package storage

import (
	"context"

	"api-gateway/internal/application/user"
)

type Handler struct {
	client user.Client
}

func NewHandler(client user.Client) *Handler {
	return &Handler{
		client: client,
	}
}

func (h *Handler) Handle(ctx context.Context, userID string) (*user.Storage, error) {
	return h.client.GetStorage(ctx, userID)
}
