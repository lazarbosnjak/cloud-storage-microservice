package grpc

import (
	"context"
	pb "proto/api-gateway"
	"time"

	userping "api-gateway/internal/application/user/ping"
)

func (s *Server) Ping(_ context.Context, in *pb.PingRequest) (*pb.PingReply, error) {
	time.Sleep(time.Second * 5)
	return &pb.PingReply{Message: "hello from api-gateway"}, nil
}

func (s *Server) PingUser(ctx context.Context, in *pb.PingRequest) (*pb.PingReply, error) {
	res, err := s.Handlers.UserPingHandler.Handle(ctx, userping.Query{})
	if err != nil {
		return nil, err
	}

	return &pb.PingReply{Message: res.Message}, nil
}
