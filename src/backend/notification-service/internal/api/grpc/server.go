package grpc

import (
	"context"
	"time"

	pb "proto/notification"
)

type Server struct {
	pb.UnimplementedNotificationServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingReply, error) {
	time.Sleep(time.Second * 5)
	return &pb.PingReply{Message: "hello from notification-service"}, nil
}
