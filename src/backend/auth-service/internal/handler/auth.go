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
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, tokenID string, newPassword string, repeatedPassword string) error
	RequestMagicLink(ctx context.Context, email string) error
	MagicLinkSignIn(ctx context.Context, tokenID string) (service.SignInResult, error)
}

type AuthHandler struct {
	pb.AuthServer
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
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
		switch {
		case errors.Is(err, domain.ErrEmailTaken):
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		case errors.Is(err, domain.ErrInvalidEmail), errors.Is(err, domain.ErrPassword):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, domain.ErrFirstName), errors.Is(err, domain.ErrLastName):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if _, ok := status.FromError(err); ok {
			return nil, err
		}

		slog.Error("register credential", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.RegisterResponse{}, nil
}

func (h *AuthHandler) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.SignInResponse, error) {
	res, err := h.service.SignIn(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}

		if errors.Is(err, domain.ErrAccountLocked) {
			return nil, status.Error(codes.ResourceExhausted, "account temporarly locked, try again later")
		}
		slog.ErrorContext(ctx, "sign in failed", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.SignInResponse{
		OtpRequired:        res.OtpRequired,
		OtpChallenge:       res.OtpChallenge,
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
			return nil, status.Error(codes.InvalidArgument, "new password does not meet requirements")
		case errors.Is(err, domain.ErrAccountLocked):
			return nil, status.Error(codes.ResourceExhausted, "account temporarily locked, try again later")
		case errors.Is(err, domain.ErrFirstName):
			return nil, status.Error(codes.InvalidArgument, "first name is not valid")
		case errors.Is(err, domain.ErrLastName):
			return nil, status.Error(codes.InvalidArgument, "last name is not valid")
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

func (h *AuthHandler) ForgotPassword(
	ctx context.Context,
	req *pb.ForgotPasswordRequest,
) (*pb.ForgotPasswordResponse, error) {
	err := h.service.ForgotPassword(ctx, req.GetEmail())
	if err != nil {
		// Swallowed on purpose: the response must never reveal whether the
		// email exists or why the flow failed internally.
		slog.ErrorContext(ctx, "forgot password failed", "error", err)
	}

	return &pb.ForgotPasswordResponse{}, nil
}

func (h *AuthHandler) ResetPassword(
	ctx context.Context,
	req *pb.ResetPasswordRequest,
) (*pb.ResetPasswordResponse, error) {
	err := h.service.ResetPassword(ctx, req.GetTokenId(), req.GetNewPassword(), req.GetRepeatedPassword())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPasswordMismatch), errors.Is(err, domain.ErrPassword):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, domain.ErrVerificationTokenNotFound),
			errors.Is(err, domain.ErrInvalidVerificationToken),
			errors.Is(err, domain.ErrVerificationTokenExpired):
			return nil, status.Error(codes.InvalidArgument, "invalid or expired reset link")
		default:
			slog.ErrorContext(ctx, "reset password failed", "error", err)
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &pb.ResetPasswordResponse{}, nil
}

func (h *AuthHandler) VerifyOtp(ctx context.Context, req *pb.VerifyOtpRequest) (*pb.VerifyOtpResponse, error) {
	res, err := h.service.VerifyOtp(ctx, req.GetOtpChallenge(), req.GetCode())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOtpInvalid):
			return nil, status.Error(codes.Unauthenticated, "invalid verification code")
		case errors.Is(err, domain.ErrOtpExpired):
			return nil, status.Error(codes.DeadlineExceeded, "verification code expired, sign in again")
		case errors.Is(err, domain.ErrOtpTooMany):
			return nil, status.Error(codes.ResourceExhausted, "too many incorrect codes, sign up again")
		case errors.Is(err, domain.ErrConcurrentRefresh):
			return nil, status.Error(codes.Unauthenticated, "verification code already used")
		default:
			slog.ErrorContext(ctx, "verify otp failed", "error", err)
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &pb.VerifyOtpResponse{
		AccessToken:        res.AccessToken,
		ExpiresInSeconds:   int64(res.ExpiresIn.Seconds()),
		MustChangePassword: res.MustChangePassword,
		RefreshToken:       res.RefreshToken,
	}, nil

}

func (h *AuthHandler) RequestMagicLink(ctx context.Context, req *pb.RequestMagicLinkRequest) (*pb.RequestMagicLinkResponse, error) {
	err := h.service.RequestMagicLink(ctx, req.GetEmail())
	if err != nil {
		slog.ErrorContext(ctx, "magic link request failed", "error", err)
	}

	return &pb.RequestMagicLinkResponse{}, nil
}

func (h *AuthHandler) MagicLinkSignIn(ctx context.Context, req *pb.MagicLinkSignInRequest) (*pb.MagicLinkSignInResponse, error) {
	res, err := h.service.MagicLinkSignIn(ctx, req.GetTokenId())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidVerificationToken):
			return nil, status.Error(codes.Unauthenticated, "invalid sign-in link")
		case errors.Is(err, domain.ErrVerificationTokenExpired):
			return nil, status.Error(codes.DeadlineExceeded, "sign-in link expired, request a new one")
		case errors.Is(err, domain.ErrVerificationTokenNotFound):
			return nil, status.Error(codes.NotFound, "sign-in link not found")
		default:
			slog.ErrorContext(ctx, "magic link sign in failed", "error", err)
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &pb.MagicLinkSignInResponse{
		AccessToken:      res.AccessToken,
		ExpiresInSeconds: int64(res.ExpiresIn.Seconds()),
		RefreshToken:     res.RefreshToken,
	}, nil
}
