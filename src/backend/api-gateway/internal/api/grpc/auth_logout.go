package grpc

import (
	"context"
	pb "proto/api-gateway"

	"api-gateway/internal/application/auth/logout"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Server) AuthSignOut(
	ctx context.Context,
	_ *pb.AuthSignOutRequest,
) (*pb.AuthSignOutResponse, error) {
	var refreshToken string
	md, _ := metadata.FromIncomingContext(ctx)
	if v := md.Get("cookie-refresh_token"); len(v) > 0 {
		refreshToken = v[0]
	} else {
		return nil, status.Error(codes.InvalidArgument, "refresh token missing")
	}

	req := logout.Request{
		RefreshToken: refreshToken,
	}

	err := s.Handlers.AuthSignOutHandler.Handle(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.AuthSignOutResponse{}, nil
}
