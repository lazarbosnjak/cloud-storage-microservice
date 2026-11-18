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

type Client interface {
	Register(context.Context, RegisterUserDTO) error
}
