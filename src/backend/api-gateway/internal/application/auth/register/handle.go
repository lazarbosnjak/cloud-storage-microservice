package register

import (
	"context"

	"api-gateway/internal/application/auth"
)

type Command struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
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
	user := auth.RegisterUserDTO{
		Email:     command.Email,
		Password:  command.Password,
		FirstName: command.FirstName,
		LastName:  command.LastName,
	}

	err := h.client.Register(ctx, user)
	if err != nil {
		return err
	}

	return nil
}
