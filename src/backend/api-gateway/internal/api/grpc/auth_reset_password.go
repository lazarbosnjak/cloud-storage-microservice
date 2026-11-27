package grpc

import (
	"context"
	authPb "proto/auth"

	resetPassword "api-gateway/internal/application/auth/reset_password"
)

func (s *Server) AuthResetPassword(ctx context.Context, in *authPb.ResetPasswordRequest) (*authPb.ResetPasswordResponse, error) {
	command := resetPassword.Command{
		TokenID:          in.GetTokenId(),
		NewPassword:      in.GetNewPassword(),
		RepeatedPassword: in.GetRepeatedPassword(),
	}

	err := s.Handlers.AuthResetPasswordHandler.Handle(ctx, command)
	if err != nil {
		return nil, err
	}

	return &authPb.ResetPasswordResponse{}, nil
}
