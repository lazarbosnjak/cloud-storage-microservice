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
	AccessToken      string
	ExpiresInSeconds int
	RefreshToken     string
}

type Client interface {
	Register(context.Context, RegisterUserDTO) error
	Login(context.Context, LoginDTO) (*LoginResponseDTO, error)
}
