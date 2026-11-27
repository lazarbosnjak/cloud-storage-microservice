package change_password

import (
	"context"

	"api-gateway/internal/application/auth"
)

type Command struct {
	Email           string
	CurrentPassword string
	NewPassword     string
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
	dto := auth.ChangePasswordDTO{
		Email:           command.Email,
		CurrentPassword: command.CurrentPassword,
		NewPassword:      command.NewPassword,
	}

	err := h.client.ChangePassword(ctx, dto)
	if err != nil {
		return err
	}

	return nil
}
