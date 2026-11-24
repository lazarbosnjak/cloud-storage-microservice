package grpc

import (
	"context"
	pb "proto/api-gateway"
	authPb "proto/auth"

	"api-gateway/internal/application/auth/login"
)

func (s *Server) AuthSignIn(ctx context.Context, in *authPb.SignInRequest) (*pb.AuthSignInResponse, error) {
	query := login.Query{
		Username: in.GetEmail(),
		Password: in.GetPassword(),
	}

	handleRes, err := s.Handlers.AuthLoginHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	return &pb.AuthSignInResponse{
		AccessToken:      handleRes.AccessToken,
		ExpiresInSeconds: handleRes.ExpiresInSeconds,
		OtpRequired:      handleRes.OtpRequired,
		OtpChallenge:     handleRes.OtpChallenge,
	}, nil
}
