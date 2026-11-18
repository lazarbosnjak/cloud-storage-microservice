package auth

import (
	"context"
)

type RegisterUserDTO struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type Client interface {
	Register(context.Context, RegisterUserDTO) error
}
