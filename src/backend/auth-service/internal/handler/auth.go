package handler

import (
	"context"
	"log/slog"

	auth "proto/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Ping(ctx context.Context) (string, error)
	Register(ctx context.Context, email string, password string) (string, error)
}

type AuthHandler struct {
	auth.UnimplementedAuthServer
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Ping(ctx context.Context, req *auth.PingRequest) (*auth.PingReply, error) {
	message, err := h.service.Ping(ctx)
	if err != nil {
		slog.Error("unable to ping")
		return nil, status.Errorf(codes.Internal, "unable to ping")
	}

	return &auth.PingReply{Message: message}, nil
}

func (h *AuthHandler) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	id, _ := h.service.Register(ctx, req.GetEmail(), req.GetPassword())

	return &auth.RegisterResponse{UserId: id}, nil
}
