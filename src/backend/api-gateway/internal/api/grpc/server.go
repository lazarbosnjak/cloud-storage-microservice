package grpc

import (
	pb "proto/api-gateway"

	authlogin "api-gateway/internal/application/auth/login"
	authrefresh "api-gateway/internal/application/auth/refresh"
	authregister "api-gateway/internal/application/auth/register"
	filehandler "api-gateway/internal/application/file/file"
	folderhandler "api-gateway/internal/application/file/folder"
	userping "api-gateway/internal/application/user/ping"
)

type Server struct {
	pb.UnimplementedApiGatewayServer
	Handlers *Handlers
}

type Handlers struct {
	UserPingHandler     *userping.Handler
	AuthRegisterHandler *authregister.Handler
	AuthLoginHandler    *authlogin.Handler
	AuthRefreshHandler  *authrefresh.Handler
	FileHandler         *filehandler.Handler
	FolderHandler       *folderhandler.Handler
}

func NewServer(handlers *Handlers) *Server {
	return &Server{
		Handlers: handlers,
	}
}
