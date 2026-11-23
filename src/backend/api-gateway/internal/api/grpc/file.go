package grpc

import (
	"context"

	pb "proto/api-gateway"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func ownerIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing identity")
	}

	values := md.Get("x-user-id")
	if len(values) != 1 || values[0] == "" {
		return "", status.Error(codes.Unauthenticated, "invalid identity")
	}

	return values[0], nil
}

func (s *Server) GetFile(
	ctx context.Context,
	req *pb.GatewayGetFileRequest,
) (*pb.GatewayFileMetadataReply, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req == nil || req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "file id is required")
	}
	result, err := s.Handlers.FileHandler.GetFile(
		ctx,
		req.GetId(),
		ownerID,
	)
	if err != nil {
		return nil, err
	}

	return &pb.GatewayFileMetadataReply{
		Id:          result.ID,
		Name:        result.Name,
		Size:        result.Size,
		ContentType: result.ContentType,
		ParentId:    result.ParentID,
		OwnerId:     result.OwnerID,
		CreatedAt:   result.CreatedAt,
		UpdatedAt:   result.UpdatedAt,
	}, nil
}

func (s *Server) GetFolder(
	ctx context.Context,
	req *pb.GatewayGetFolderRequest,
) (*pb.GatewayFolderMetadataReply, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req == nil || req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "folder id is required")
	}
	result, err := s.Handlers.FolderHandler.GetFolder(
		ctx,
		req.GetId(),
		ownerID,
	)
	if err != nil {
		return nil, err
	}

	return &pb.GatewayFolderMetadataReply{
		Id:        result.ID,
		Name:      result.Name,
		ParentId:  result.ParentID,
		OwnerId:   result.OwnerID,
		CreatedAt: result.CreatedAt,
		UpdatedAt: result.UpdatedAt,
	}, nil
}
