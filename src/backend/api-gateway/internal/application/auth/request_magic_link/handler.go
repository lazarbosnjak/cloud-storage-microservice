package requestmagiclink

import (
	"api-gateway/internal/application/auth"
	"context"
)

type Command struct {
	Email string
}

type Handler struct {
	client auth.Client
}

func NewHandler(client auth.Client) *Handler {
	return &Handler{
		client: client,
	}
}

func (h *Handler) Handle(ctx context.Context, command Command) error {
	return h.client.RequestMagicLink(ctx, auth.RequestMagicLinkDTO{
		Email: command.Email,
	})
}
