package grpc

import (
	"context"
	pb "proto/user"

	"user-service/internal/application/ping"
)

func (s *Server) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingReply, error) {
	handler := ping.NewHandler()
	message := handler.Handle(ctx)

	reply := &pb.PingReply{
		Message: message,
	}

	return reply, nil
}
