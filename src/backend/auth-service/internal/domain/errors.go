package domain

import "errors"

var (
	ErrNotFound           = errors.New("credential not found")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrAccountLocked      = errors.New("account temporarly locked")
)
