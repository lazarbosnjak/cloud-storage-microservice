package grpc

import (
	"context"
	pb "proto/api-gateway"

	"api-gateway/internal/application/auth/refresh"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Server) AuthRefresh(ctx context.Context, _ *pb.AuthRefreshRequest) (*pb.AuthRefreshResponse, error) {
	var refreshToken string
	md, _ := metadata.FromIncomingContext(ctx)
	if v := md.Get("cookie-refresh_token"); len(v) > 0 {
		refreshToken = v[0]
	} else {
		return nil, status.Error(codes.InvalidArgument, "refresh token missing")
	}

	req := refresh.Request{
		RefreshToken: refreshToken,
	}

	res, err := s.Handlers.AuthRefreshHandler.Handle(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.AuthRefreshResponse{
		AccessToken:      res.AccessToken,
		ExpiresInSeconds: int64(res.ExpiresInSeconds),
	}, err
}
