package grpc

import (
	"context"
	pb "proto/api-gateway"

	verifyregistration "api-gateway/internal/application/auth/verify_registration"
)

func (s *Server) AuthVerifyRegistration(
	ctx context.Context,
	in *pb.AuthVerifyRegistrationRequest,
) (*pb.AuthVerifyRegistrationResponse, error) {
	req := verifyregistration.Request{
		VerificationToken: in.GetVerificationToken(),
	}

	err := s.Handlers.AuthVerifyRegistrationHandler.Handle(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.AuthVerifyRegistrationResponse{}, nil
}
