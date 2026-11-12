package ping

import "context"

type UserPinger interface {
	Ping(ctx context.Context) (string, error)
}

type Result struct {
	Message string
}

type Handler struct {
	userPinger UserPinger
}

func NewHandler(userPinger UserPinger) *Handler {
	return &Handler{
		userPinger: userPinger,
	}
}

func (h *Handler) Handle(
	ctx context.Context,
	_ Query,
) (Result, error) {
	message, err := h.userPinger.Ping(ctx)
	if err != nil {
		return Result{}, err
	}

	return Result{Message: message}, nil
}
