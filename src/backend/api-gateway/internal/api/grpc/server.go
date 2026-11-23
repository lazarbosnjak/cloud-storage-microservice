package grpc

import (
	pb "proto/api-gateway"

	authlogin "api-gateway/internal/application/auth/login"
	authrefresh "api-gateway/internal/application/auth/refresh"
	authregister "api-gateway/internal/application/auth/register"
	getfile "api-gateway/internal/application/file/get_file"
	getfolder "api-gateway/internal/application/file/get_folder"
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
	FileGetHandler      *getfile.Handler
	FolderGetHandler    *getfolder.Handler
}

func NewServer(handlers *Handlers) *Server {
	return &Server{
		Handlers: handlers,
	}
}
