package grpc

import (
	pb "proto/api-gateway"

	authlogin "api-gateway/internal/application/auth/login"
	authrefresh "api-gateway/internal/application/auth/refresh"
	authregister "api-gateway/internal/application/auth/register"
	authverifyregistration "api-gateway/internal/application/auth/verify_registration"
	changePassword "api-gateway/internal/application/auth/change_password"

	verifyotp "api-gateway/internal/application/auth/verify-otp"
	userping "api-gateway/internal/application/user/ping"
)

type Server struct {
	pb.UnimplementedApiGatewayServer
	Handlers *Handlers
}

type Handlers struct {
	UserPingHandler               *userping.Handler
	AuthRegisterHandler           *authregister.Handler
	AuthLoginHandler              *authlogin.Handler
	AuthRefreshHandler            *authrefresh.Handler
	AuthVerifyRegistrationHandler *authverifyregistration.Handler
	AuthVerifyOtpHandler          *verifyotp.Handler
	AuthChangePasswordHandler     *changePassword.Handler
}

func NewServer(handlers *Handlers) *Server {
	return &Server{
		Handlers: handlers,
	}
}
