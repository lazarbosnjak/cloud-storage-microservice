package ping

import "context"

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(ctx context.Context) string {
	return "hello from user-service"
}
