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

	if res.MustChangePassword {
		return &auth.LoginResponseDTO{
			AccessToken:        "",
			ExpiresInSeconds:   0,
			RefreshToken:       "",
			OtpRequired:        false,
			OtpChallenge:       "",
			MustChangePassword: true,
		}, nil
	}

	return &auth.LoginResponseDTO{
		AccessToken:        res.GetAccessToken(),
		ExpiresInSeconds:   int(res.GetExpiresInSeconds()),
		RefreshToken:       res.GetRefreshToken(),
		OtpRequired:        res.GetOtpRequired(),
		OtpChallenge:       res.GetOtpChallenge(),
		MustChangePassword: false,
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

func (c *GRPCClient) VerifyRegistration(ctx context.Context, dto auth.VerifyRegistrationRequestDTO) error {
	req := &authPb.VerifyRegistrationRequest{
		TokenId: dto.VerificationToken,
	}

	_, err := c.client.VerifyRegistration(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (c *GRPCClient) VerifyOtp(ctx context.Context, dto auth.VerifyOtpDTO) (*auth.VerifyOtpResponseDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	res, err := c.client.VerifyOtp(ctx, &authPb.VerifyOtpRequest{
		OtpChallenge: dto.ChallengeID,
		Code:         dto.Code,
	})
	if err != nil {
		return nil, err
	}

	return &auth.VerifyOtpResponseDTO{
		AccessToken:      res.GetAccessToken(),
		ExpiresInSeconds: int(res.GetExpiresInSeconds()),
		RefreshToken:     res.GetRefreshToken(),
	}, nil
}

func (c *GRPCClient) ChangePassword(ctx context.Context, dto auth.ChangePasswordDTO) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := c.client.ChangePassword(ctx, &authPb.ChangePasswordRequest{
		Email:           dto.Email,
		CurrentPassword: dto.CurrentPassword,
		NewPassword:     dto.NewPassword,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *GRPCClient) ForgotPassword(ctx context.Context, dto auth.ForgotPasswordDTO) error {
	_, err := c.client.ForgotPassword(ctx, &authPb.ForgotPasswordRequest{
		Email: dto.Email,
	})
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}

	return nil
}

func (c *GRPCClient) ResetPassword(ctx context.Context, dto auth.ResetPasswordDTO) error {
	_, err := c.client.ResetPassword(ctx, &authPb.ResetPasswordRequest{
		TokenId:          dto.TokenID,
		NewPassword:      dto.NewPassword,
		RepeatedPassword: dto.RepeatedPassword,
	})
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}

	return nil
}

func (c *GRPCClient) SignOut(
	ctx context.Context,
	dto auth.SignOutRequestDTO,
) error {
	req := &authPb.SignOutRequest{
		RefreshToken: dto.RefreshToken,
	}

	_, err := c.client.SignOut(ctx, req)

	return err
}

func (c *GRPCClient) RequestMagicLink(ctx context.Context, dto auth.RequestMagicLinkDTO) error {
	_, err := c.client.RequestMagicLink(ctx, &authPb.RequestMagicLinkRequest{
		Email: dto.Email,
	})
	if err != nil {
		c.logger.Error(err.Error())
		return err
	}

	return nil
}

func (c *GRPCClient) MagicLinkSignIn(ctx context.Context, dto auth.MagicLinkSignInDTO) (*auth.MagicLinkSignInResponseDTO, error) {
	res, err := c.client.MagicLinkSignIn(ctx, &authPb.MagicLinkSignInRequest{
		TokenId: dto.TokenID,
	})
	if err != nil {
		c.logger.Error(err.Error())
		return nil, err
	}

	return &auth.MagicLinkSignInResponseDTO{
		AccessToken:      res.GetAccessToken(),
		ExpiresInSeconds: int(res.GetExpiresInSeconds()),
		RefreshToken:     res.GetRefreshToken(),
	}, nil
}
