package grpc

import (
	"context"

	"file-service/internal/application/ping"

	pb "proto/file"
)

type Server struct {
	pb.UnimplementedFileServer
	pingHandler *ping.Handler
}

func NewServer(pingHandler *ping.Handler) *Server {
	return &Server{
		pingHandler: pingHandler,
	}
}

func (s *Server) Ping(
	ctx context.Context,
	_ *pb.PingRequest,
) (*pb.PingReply, error) {
	message, err := s.pingHandler.Handle(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.PingReply{
		Message: message,
	}, nil
}
