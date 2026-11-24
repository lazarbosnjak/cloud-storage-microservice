package verifyregistration

import (
	"context"

	"api-gateway/internal/application/auth"
)

type Handler struct {
	authClient auth.Client
}

func NewHandler(authClient auth.Client) *Handler {
	return &Handler{
		authClient: authClient,
	}
}

type Request struct {
	VerificationToken string
}

func (h *Handler) Handle(ctx context.Context, req Request) error {
	dto := auth.VerifyRegistrationRequestDTO{
		VerificationToken: req.VerificationToken,
	}

	err := h.authClient.VerifyRegistration(ctx, dto)
	if err != nil {
		return err
	}

	return nil
}
