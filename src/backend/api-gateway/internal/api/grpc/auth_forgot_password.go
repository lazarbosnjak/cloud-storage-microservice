package grpc

import (
	"context"
	authPb "proto/auth"

	forgotPassword "api-gateway/internal/application/auth/forgot_password"
)

func (s *Server) AuthForgotPassword(ctx context.Context, in *authPb.ForgotPasswordRequest) (*authPb.ForgotPasswordResponse, error) {
	command := forgotPassword.Command{
		Email: in.GetEmail(),
	}

	err := s.Handlers.AuthForgotPasswordHandler.Handle(ctx, command)
	if err != nil {
		return nil, err
	}

	return &authPb.ForgotPasswordResponse{}, nil
}
