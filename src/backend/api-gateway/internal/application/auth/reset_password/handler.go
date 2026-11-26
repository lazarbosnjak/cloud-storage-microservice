package reset_password

import (
	"api-gateway/internal/application/auth"
	"context"
)

type Command struct {
	TokenID          string
	NewPassword      string
	RepeatedPassword string
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
	return h.client.ResetPassword(ctx, auth.ResetPasswordDTO{
		TokenID:          command.TokenID,
		NewPassword:      command.NewPassword,
		RepeatedPassword: command.RepeatedPassword,
	})
}
