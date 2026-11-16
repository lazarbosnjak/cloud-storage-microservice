package ping

import (
	"context"

	"api-gateway/internal/application/user"
)

type Result struct {
	Message string
}

type Handler struct {
	userClient user.Client
}

func NewHandler(userClient user.Client) *Handler {
	return &Handler{
		userClient: userClient,
	}
}

func (h *Handler) Handle(ctx context.Context, _ Query) (Result, error) {
	message, err := h.userClient.Ping(ctx)
	if err != nil {
		return Result{}, err
	}

	return Result{Message: message}, nil
}
