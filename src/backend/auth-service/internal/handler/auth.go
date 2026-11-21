package handler

import (
	"auth-service/internal/domain"
	"auth-service/internal/service"
	"context"
	"errors"
	"log/slog"

	pb "proto/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Ping(ctx context.Context) (string, error)
	Register(ctx context.Context, email string, password string) (string, error)
	SignIn(ctx context.Context, email string, password string) (service.SignInResult, error)
}

type AuthHandler struct {
	pb.AuthServer
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingReply, error) {
	message, err := h.service.Ping(ctx)
	if err != nil {
		slog.Error("unable to ping")
		return nil, status.Errorf(codes.Internal, "unable to ping")
	}

	return &pb.PingReply{Message: message}, nil
}

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	id, _ := h.service.Register(ctx, req.GetEmail(), req.GetPassword())

	return &pb.RegisterResponse{UserId: id}, nil
}

func (h *AuthHandler) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.SignInResponse, error) {
	res, err := h.service.SignIn(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}

		// We are deliberately avoiding throwing this error in service layer since it can tell the hacker that account exists.
		if errors.Is(err, domain.ErrAccountLocked) {
			return nil, status.Error(codes.ResourceExhausted, "account temporarly locked, try again later")
		}
		slog.ErrorContext(ctx, "sign in failed", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.SignInResponse{
		AccessToken:        res.AccessToken,
		ExpiresInSeconds:   int64(res.ExpiresIn.Seconds()),
		MustChangePassword: res.MustChangePassword,
		RefreshToken:       res.RefreshToken,
	}, nil

}
