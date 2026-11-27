package grpc

import (
	"context"

	pb "proto/api-gateway"
)

func (s *Server) GetStorage(ctx context.Context, _ *pb.GatewayGetStorageRequest) (*pb.GatewayStorageReply, error) {
	userID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.Handlers.UserStorageHandler.Handle(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &pb.GatewayStorageReply{
		Quota:     result.Quota,
		Used:      result.Used,
		Available: result.Available,
	}, nil
}
