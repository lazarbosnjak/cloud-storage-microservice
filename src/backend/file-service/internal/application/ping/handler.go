package ping

import (
	"context"
	"time"
)

type Handler struct {
	delay time.Duration
}

func NewHandler(delay time.Duration) *Handler {
	return &Handler{
		delay: delay,
	}
}

func (h *Handler) Handle(ctx context.Context) (string, error) {
	timer := time.NewTimer(h.delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return "hello from file-service", nil

	case <-ctx.Done():
		return "", ctx.Err()
	}
}
