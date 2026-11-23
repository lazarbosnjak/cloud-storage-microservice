package auth

import (
	"context"
	"log/slog"
	authPb "proto/auth"
	"time"

	"api-gateway/internal/application/auth"
)

type GRPCClient struct {
	client authPb.AuthClient
	logger *slog.Logger
}

func NewGRPCClient(client authPb.AuthClient, logger *slog.Logger) *GRPCClient {
	return &GRPCClient{
		client: client,
		logger: logger,
	}
}

func (c *GRPCClient) Ping(ctx context.Context) (string, error) {
	reply, err := c.client.Ping(ctx, &authPb.PingRequest{})
	if err != nil {
		c.logger.Error(err.Error())
		return "", err
	}

	return reply.Message, nil
}

func (c *GRPCClient) Register(ctx context.Context, dto auth.RegisterUserDTO) error {
	req := &authPb.RegisterRequest{
		Email:            dto.Email,
		Password:         dto.Password,
		RepeatedPassword: dto.RepeatedPassword,
		FirstName:        dto.FirstName,
		LastName:         dto.LastName,
	}

	_, err := c.client.Register(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (c *GRPCClient) Login(ctx context.Context, dto auth.LoginDTO) (*auth.LoginResponseDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req := &authPb.SignInRequest{
		Email:    dto.Email,
		Password: dto.Password,
	}

	res, err := c.client.SignIn(ctx, req)
	if err != nil {
		return nil, err
	}

	return &auth.LoginResponseDTO{
		AccessToken:      res.GetAccessToken(),
		ExpiresInSeconds: int(res.GetExpiresInSeconds()),
		RefreshToken:     res.GetRefreshToken(),
	}, nil
}

func (c *GRPCClient) Refresh(ctx context.Context, dto auth.RefreshRequestDTO) (*auth.RefreshResponseDTO, error) {
	req := &authPb.RefreshRequest{
		RefreshToken: dto.RefreshToken,
	}

	res, err := c.client.Refresh(ctx, req)
	if err != nil {
		return nil, err
	}

	return &auth.RefreshResponseDTO{
		AccessToken:      res.GetAccessToken(),
		RefreshToken:     res.GetRefreshToken(),
		ExpiresInSeconds: int(res.GetExpiresInSeconds()),
	}, nil
}
