package refresh

import (
	"context"
	"net/http"
	"time"

	"api-gateway/internal/application/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Request struct {
	RefreshToken string
}

type Response struct {
	AccessToken      string
	ExpiresInSeconds int
}

type Handler struct {
	client auth.Client
}

func NewHandler(client auth.Client) *Handler {
	return &Handler{
		client: client,
	}
}

func (h *Handler) Handle(ctx context.Context, request Request) (*Response, error) {
	req := auth.RefreshRequestDTO{
		RefreshToken: request.RefreshToken,
	}
	res, err := h.client.Refresh(ctx, req)
	if err != nil {
		return nil, err
	}

	// TODO: Move cookie creation to another file

	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    res.RefreshToken,
		Path:     "/api/auth",
		MaxAge:   int(time.Hour * 24 * 7),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode, // TODO: research this
	}

	if err := grpc.SetHeader(ctx, metadata.Pairs("set-cookie", cookie.String())); err != nil {
		return nil, err
	}

	return &Response{
		AccessToken:      res.AccessToken,
		ExpiresInSeconds: res.ExpiresInSeconds,
	}, nil
}
