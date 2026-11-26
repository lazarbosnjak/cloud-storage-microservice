package grpc

import (
	pb "proto/api-gateway"

	changePassword "api-gateway/internal/application/auth/change_password"
	authlogin "api-gateway/internal/application/auth/login"
	authrefresh "api-gateway/internal/application/auth/refresh"
	authregister "api-gateway/internal/application/auth/register"
	verifyotp "api-gateway/internal/application/auth/verify-otp"
	authverifyregistration "api-gateway/internal/application/auth/verify_registration"
	notificationclient "api-gateway/internal/infrastructure/clients/notification"

	filehandler "api-gateway/internal/application/file/file"
	folderhandler "api-gateway/internal/application/file/folder"
	userping "api-gateway/internal/application/user/ping"
	userstorage "api-gateway/internal/application/user/storage"

	forgotPassword "api-gateway/internal/application/auth/forgot_password"
	resetPassword "api-gateway/internal/application/auth/reset_password"
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
	FileHandler                   *filehandler.Handler
	FolderHandler                 *folderhandler.Handler
	UserStorageHandler            *userstorage.Handler
	AuthForgotPasswordHandler     *forgotPassword.Handler
	AuthResetPasswordHandler      *resetPassword.Handler

	// TODO: will figure out better way to handle notifications...
	// For now they skip the application layer and just get forwarded to frontend
	NotificationClient *notificationclient.GRPCClient
}

func NewServer(handlers *Handlers) *Server {
	return &Server{
		Handlers: handlers,
	}
}
