package logout

import (
	"context"
	"fmt"
	"net/http"

	"api-gateway/internal/application/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Handler struct {
	client auth.Client
}

func NewHandler(client auth.Client) *Handler {
	return &Handler{
		client: client,
	}
}

type Request struct {
	RefreshToken string
}

func (h *Handler) Handle(ctx context.Context, req Request) error {
	dto := auth.SignOutRequestDTO{
		RefreshToken: req.RefreshToken,
	}

	err := h.client.SignOut(ctx, dto)
	if err != nil {
		return fmt.Errorf("error handling login: %w", err)
	}

	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/auth",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode, // TODO: research this
	}

	if err := grpc.SetHeader(ctx, metadata.Pairs("set-cookie", cookie.String())); err != nil {
		return err
	}

	return nil
}
