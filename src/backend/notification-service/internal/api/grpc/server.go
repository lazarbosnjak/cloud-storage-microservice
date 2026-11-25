package grpc

import (
	"context"
	"time"

	"notification-service/internal/infrastructure/db/cassandra"

	pb "proto/notification"
)

type Server struct {
	pb.UnimplementedNotificationServer

	NotificationRepository *cassandra.NotificationsRepository // TEST:
}

func NewServer(notifRepo *cassandra.NotificationsRepository) *Server {
	return &Server{
		NotificationRepository: notifRepo,
	}
}

func (s *Server) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingReply, error) {
	time.Sleep(time.Second * 5)

	return &pb.PingReply{Message: "hello from notification-service"}, nil
}
