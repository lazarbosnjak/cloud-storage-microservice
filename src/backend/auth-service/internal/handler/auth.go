package handler

import (
	"auth-service/internal/service"
	"context"
	"errors"
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
		return nil, status.Errorf(codes.Internal, "unable to ping")
	}

	return &auth.PingReply{Message: message}, nil
}

func (h *AuthHandler) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	id, err := h.service.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			return nil, status.Errorf(codes.InvalidArgument, "invalid email")
		case errors.Is(err, service.ErrEmailTaken):
			return nil, status.Errorf(codes.AlreadyExists, "email already registered")
		case errors.Is(err, service.ErrPassword):
			return nil, status.Errorf(codes.InvalidArgument, "password does not meet requirements")
		default:
			slog.ErrorContext(ctx, "register failed", "error", err)
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &auth.RegisterResponse{UserId: id}, nil
}
