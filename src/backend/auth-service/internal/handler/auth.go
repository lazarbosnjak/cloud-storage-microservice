package handler

import (
	"context"
	"errors"
	"log/slog"

	"auth-service/internal/domain"
	"auth-service/internal/service"

	pb "proto/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Ping(ctx context.Context) (string, error)
	Register(ctx context.Context, user service.RegisterUserDTO) (string, error)
	SignIn(ctx context.Context, email string, password string) (service.SignInResult, error)
	ChangePassword(ctx context.Context, email string, currentPassword string, newPassword string) error
	Refresh(ctx context.Context, refreshToken string) (service.RefreshResult, error)
	SignOut(ctx context.Context, refreshToken string) error
	VerifyRegistration(ctx context.Context, tokenID string) error
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
	user := service.RegisterUserDTO{
		Email:            req.GetEmail(),
		Password:         req.GetPassword(),
		RepeatedPassword: req.GetRepeatedPassword(),
		FirstName:        req.GetFirstName(),
		LastName:         req.GetLastName(),
	}

	_, err := h.service.Register(ctx, user)
	if err != nil {
		if errors.Is(err, domain.ErrEmailTaken) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		slog.Error("register credential", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.RegisterResponse{}, nil
}

func (h *AuthHandler) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.SignInResponse, error) {
	res, err := h.service.SignIn(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) || errors.Is(err, domain.ErrAccountLocked) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}

		// We are deliberately avoiding throwing this error since it can tell the hacker that account exists.
		/*if errors.Is(err, domain.ErrAccountLocked) {
			return nil, status.Error(codes.ResourceExhausted, "account temporarly locked, try again later")
		}*/
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

func (h *AuthHandler) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	err := h.service.ChangePassword(ctx, req.GetEmail(), req.GetCurrentPassword(), req.GetNewPassword())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		case errors.Is(err, domain.ErrPasswordUnchanged):
			return nil, status.Error(codes.InvalidArgument, "new password must differ from the current one")
		case errors.Is(err, domain.ErrPassword):
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		case errors.Is(err, domain.ErrAccountLocked):
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		/*
			case errors.Is(err, domain.ErrPasswordUnchanged):
				return nil, status.Error(codes.InvalidArgument, "new password must differ from the current one")
			case errors.Is(err, domain.ErrPassword):
				return nil, status.Error(codes.InvalidArgument, "new password does not meet requirements")
			case errors.Is(err, domain.ErrAccountLocked):
				return nil, status.Error(codes.ResourceExhausted, "account temporarily locked, try again later")
		*/
		default:
			slog.ErrorContext(ctx, "change password failed", "error", err)
			return nil, status.Error(codes.Internal, "internal error")

		}
	}

	return &pb.ChangePasswordResponse{}, nil
}

func (h *AuthHandler) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	res, err := h.service.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidToken) {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		slog.ErrorContext(ctx, "refresh failed", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &pb.RefreshResponse{
		AccessToken:      res.AccessToken,
		RefreshToken:     res.RefreshToken,
		ExpiresInSeconds: int64(res.ExpiresIn.Seconds()),
	}, nil
}

func (h *AuthHandler) SignOut(ctx context.Context, req *pb.SignOutRequest) (*pb.SignOutResponse, error) {
	if err := h.service.SignOut(ctx, req.GetRefreshToken()); err != nil {
		slog.ErrorContext(ctx, "sign our failed", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &pb.SignOutResponse{}, nil
}

func (h *AuthHandler) VerifyRegistration(
	ctx context.Context,
	req *pb.VerifyRegistrationRequest,
) (*pb.VerifyRegistrationResponse, error) {
	err := h.service.VerifyRegistration(ctx, req.TokenId)

	switch {
	case errors.Is(err, domain.ErrVerificationTokenNotFound):
		return nil, status.Error(codes.NotFound, domain.ErrVerificationTokenNotFound.Error())
	case err != nil:
		return nil, err
	}

	return &pb.VerifyRegistrationResponse{}, nil
}
