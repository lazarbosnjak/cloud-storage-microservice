package grpc

import (
	"context"
	pb "proto/api-gateway"
	notifPb "proto/notification"
)

func (s *Server) NotificationsGetAllByUser(
	ctx context.Context,
	_ *pb.NotificationsGetAllByUserRequest,
) (*notifPb.GetAllByUserResponse, error) {
	userID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	req := &notifPb.GetAllByUserRequest{
		UserId: userID,
	}

	res, err := s.Handlers.NotificationClient.GetAllByUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
