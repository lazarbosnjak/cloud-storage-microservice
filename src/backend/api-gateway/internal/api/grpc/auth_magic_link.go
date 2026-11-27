package grpc

import (
	"context"

	pb "proto/api-gateway"
	authPb "proto/auth"

	magiclinksignin "api-gateway/internal/application/auth/magic_link_sign_in"
	requestmagiclink "api-gateway/internal/application/auth/request_magic_link"
)

func (s *Server) AuthRequestMagicLink(ctx context.Context, in *authPb.RequestMagicLinkRequest) (*authPb.RequestMagicLinkResponse, error) {
	command := requestmagiclink.Command{
		Email: in.GetEmail(),
	}

	err := s.Handlers.AuthRequestMagicLinkHandler.Handle(ctx, command)
	if err != nil {
		return nil, err
	}

	return &authPb.RequestMagicLinkResponse{}, nil
}

func (s *Server) AuthMagicLinkSignIn(ctx context.Context, in *authPb.MagicLinkSignInRequest) (*pb.AuthMagicLinkSignInResponse, error) {
	command := magiclinksignin.Command{
		TokenID: in.GetTokenId(),
	}

	handleRes, err := s.Handlers.AuthMagicLinkSignInHandler.Handle(ctx, command)
	if err != nil {
		return nil, err
	}

	return &pb.AuthMagicLinkSignInResponse{
		AccessToken:      handleRes.AccessToken,
		ExpiresInSeconds: handleRes.ExpiresInSeconds,
	}, nil
}
