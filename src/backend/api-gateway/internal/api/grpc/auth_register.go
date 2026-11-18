package grpc

import (
	"context"
	"proto/auth"

	"api-gateway/internal/application/auth/register"
)

func (s *Server) AuthRegister(ctx context.Context, in *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	command := register.Command{
		Email:     in.Email,
		Password:  in.Password,
		FirstName: in.FirstName,
		LastName:  in.LastName,
	}

	res := &auth.RegisterResponse{}

	err := s.Handlers.AuthRegisterHandler.Handle(ctx, command)
	if err != nil {
		// TODO: I think this hsould be the message sent back to user, so
		// format the response to have a proper message
		return res, err
	}

	return res, nil
}
