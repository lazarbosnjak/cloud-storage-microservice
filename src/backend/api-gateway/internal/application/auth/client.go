package auth

import (
	"context"
)

type RegisterUserDTO struct {
	Email            string
	Password         string
	RepeatedPassword string
	FirstName        string
	LastName         string
}

type LoginDTO struct {
	Email    string
	Password string
}

type LoginResponseDTO struct {
	AccessToken        string
	ExpiresInSeconds   int
	RefreshToken       string
	OtpRequired        bool
	OtpChallenge       string
	MustChangePassword bool
}

type RefreshRequestDTO struct {
	RefreshToken string
}

type RefreshResponseDTO struct {
	AccessToken      string
	ExpiresInSeconds int
	RefreshToken     string
}

type VerifyRegistrationRequestDTO struct {
	VerificationToken string
}

type ChangePasswordDTO struct {
	Email           string
	CurrentPassword string
	NewPassword     string
}

type VerifyOtpDTO struct {
	ChallengeID string
	Code        string
}

type VerifyOtpResponseDTO struct {
	AccessToken      string
	ExpiresInSeconds int
	RefreshToken     string
}

type ForgotPasswordDTO struct {
	Email string
}

type ResetPasswordDTO struct {
	TokenID          string
	NewPassword      string
	RepeatedPassword string
}

type Client interface {
	Register(context.Context, RegisterUserDTO) error
	Login(context.Context, LoginDTO) (*LoginResponseDTO, error)
	Refresh(context.Context, RefreshRequestDTO) (*RefreshResponseDTO, error)
	VerifyRegistration(context.Context, VerifyRegistrationRequestDTO) error
	VerifyOtp(context.Context, VerifyOtpDTO) (*VerifyOtpResponseDTO, error)
	ChangePassword(context.Context, ChangePasswordDTO) error
	ForgotPassword(context.Context, ForgotPasswordDTO) error
	ResetPassword(context.Context, ResetPasswordDTO) error
}
