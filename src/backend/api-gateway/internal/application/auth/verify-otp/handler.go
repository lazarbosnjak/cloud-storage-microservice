package verifyotp

import (
	"context"
	"net/http"
	"time"

	"api-gateway/internal/application/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Handler struct {
	authClient auth.Client
}

func NewHandler(client auth.Client) *Handler {
	return &Handler{authClient: client}
}

type Query struct {
	ChallengeID string
	Code        string
}

type Response struct {
	AccessToken      string
	ExpiresInSeconds int32
}

func (h *Handler) Handle(ctx context.Context, query Query) (*Response, error) {
	res, err := h.authClient.VerifyOtp(ctx, auth.VerifyOtpDTO{
		ChallengeID: query.ChallengeID,
		Code:        query.Code,
	})
	if err != nil {
		return nil, err
	}

	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    res.RefreshToken,
		Path:     "/api/auth/refresh",
		MaxAge:   int(time.Hour * 24 * 7),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	if err := grpc.SetHeader(ctx, metadata.Pairs("set-cookie", cookie.String())); err != nil {
		return nil, err
	}

	return &Response{
		AccessToken:      res.AccessToken,
		ExpiresInSeconds: int32(res.ExpiresInSeconds),
	}, nil
}
