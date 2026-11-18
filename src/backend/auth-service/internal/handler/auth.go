package handler

import (
	"context"
	"log/slog"

	"auth-service/internal/service"

	auth "proto/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Ping(ctx context.Context) (string, error)
	Register(ctx context.Context, user service.RegisterUserDTO) (string, error)
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
	user := service.RegisterUserDTO{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
	}
	_, err := h.service.Register(ctx, user)
	if err != nil {
		// TODO: Figure out better error message?
		// Need to test
		return nil, err
	}

	return &auth.RegisterResponse{}, nil
}
