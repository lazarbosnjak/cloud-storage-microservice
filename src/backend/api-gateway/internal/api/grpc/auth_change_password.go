package grpc

import (
	"context"
	authPb "proto/auth"

	changePassword "api-gateway/internal/application/auth/change_password"
)

func (s *Server) AuthChangePassword(ctx context.Context, in *authPb.ChangePasswordRequest) (*authPb.ChangePasswordResponse, error) {
	command := changePassword.Command{
		Email:           in.GetEmail(),
		CurrentPassword: in.GetCurrentPassword(),
		NewPassword:     in.GetNewPassword(),
	}

	err := s.Handlers.AuthChangePasswordHandler.Handle(ctx, command)
	if err != nil {
		return nil, err
	}

	return &authPb.ChangePasswordResponse{}, nil
}
