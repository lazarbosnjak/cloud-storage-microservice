package domain

import "errors"

var (
	ErrNotFound           = errors.New("credential not found")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrAccountLocked      = errors.New("account temporarly locked")
	ErrPasswordUnchanged  = errors.New("new password must differ from the current one")
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrPassword           = errors.New("password does not meet requirements")
	ErrDefault            = errors.New("there was an error")
	ErrConcurrentRefresh  = errors.New("refresh token already consumed")
	ErrPasswordMismatch   = errors.New("password and repeated password do not match")
)
