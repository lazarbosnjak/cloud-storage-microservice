package ping

import (
	"context"
	"time"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(ctx context.Context) (string, error) {
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
		return "hello from file-service", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
