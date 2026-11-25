package grpc

import (
	verifyotp "api-gateway/internal/application/auth/verify-otp"
	"context"
	pb "proto/api-gateway"
	authPb "proto/auth"
)

func (s *Server) AuthVerifyOtp(ctx context.Context, in *authPb.VerifyOtpRequest) (*pb.AuthVerifyOtpResponse, error) {
	query := verifyotp.Query{
		ChallengeID: in.GetOtpChallenge(),
		Code:        in.GetCode(),
	}

	handleRes, err := s.Handlers.AuthVerifyOtpHandler.Handle(ctx, query)
	if err != nil {
		return nil, err
	}

	return &pb.AuthVerifyOtpResponse{
		AccessToken:      handleRes.AccessToken,
		ExpiresInSeconds: handleRes.ExpiresInSeconds,
	}, nil
}
