package grpc

import (
	"context"
	"proto/auth"

	"api-gateway/internal/application/auth/register"
)

func (s *Server) AuthRegister(ctx context.Context, in *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	command := register.Command{
		Email:            in.GetEmail(),
		Password:         in.GetPassword(),
		RepeatedPassword: in.GetRepeatedPassword(),
		FirstName:        in.GetFirstName(),
		LastName:         in.GetLastName(),
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
