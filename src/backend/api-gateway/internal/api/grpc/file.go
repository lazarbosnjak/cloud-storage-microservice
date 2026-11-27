package grpc

import (
	"context"

	pb "proto/api-gateway"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	appfile "api-gateway/internal/application/file"
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

func (s *Server) CreateFile(
	ctx context.Context,
	req *pb.GatewayCreateFileRequest,
) (*pb.GatewayCreatedFileReply, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req == nil || req.GetName() == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"file name is required",
		)
	}

	result, err := s.Handlers.FileHandler.CreateFile(
		ctx,
		appfile.CreateFileInput{
			Name:        req.GetName(),
			Size:        req.GetSize(),
			ContentType: req.GetContentType(),
			ParentID:    req.GetParentId(),
			OwnerID:     ownerID,
		},
	)

	if err != nil {
		return nil, err
	}

	return &pb.GatewayCreatedFileReply{
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

func (s *Server) RenameFile(
	ctx context.Context,
	req *pb.GatewayRenameRequest,
) (*pb.GatewayEmptyReply, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req == nil || req.GetId() == "" || req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "file id and name are required")
	}

	if err := s.Handlers.FileHandler.RenameFile(
		ctx, req.GetId(), ownerID, req.GetName(),
	); err != nil {
		return nil, err
	}

	return &pb.GatewayEmptyReply{}, nil
}

func (s *Server) CreateFolder(
	ctx context.Context,
	req *pb.GatewayCreateFolderRequest,
) (*pb.GatewayCreatedFolderReply, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req == nil || req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "folder name is required")
	}

	result, err := s.Handlers.FolderHandler.CreateFolder(ctx, appfile.CreateFolderInput{
		Name:     req.GetName(),
		ParentID: req.GetParentId(),
		OwnerID:  ownerID,
	})
	if err != nil {
		return nil, err
	}

	return &pb.GatewayCreatedFolderReply{
		Id:        result.ID,
		Name:      result.Name,
		ParentId:  result.ParentID,
		OwnerId:   result.OwnerID,
		CreatedAt: result.CreatedAt,
	}, nil
}

func (s *Server) RenameFolder(
	ctx context.Context,
	req *pb.GatewayRenameRequest,
) (*pb.GatewayEmptyReply, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req == nil || req.GetId() == "" || req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "folder id and name are required")
	}

	if err := s.Handlers.FolderHandler.RenameFolder(
		ctx, req.GetId(), ownerID, req.GetName(),
	); err != nil {
		return nil, err
	}

	return &pb.GatewayEmptyReply{}, nil
}

func (s *Server) ListFolderContents(
	ctx context.Context,
	req *pb.GatewayListFolderContentsRequest,
) (*pb.GatewayFolderContentsReply, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req == nil || req.GetFolderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "folder id is required")
	}

	folderID := req.GetFolderId()

	if folderID == "root" {
		folderID = ""
	}

	result, err := s.Handlers.FolderHandler.ListFolderContents(
		ctx, folderID, ownerID,
	)
	if err != nil {
		return nil, err
	}

	reply := &pb.GatewayFolderContentsReply{}

	for _, f := range result.Files {
		reply.Files = append(reply.Files, &pb.GatewayFileMetadataReply{
			Id:          f.ID,
			Name:        f.Name,
			Size:        f.Size,
			ContentType: f.ContentType,
			ParentId:    f.ParentID,
			OwnerId:     f.OwnerID,
			CreatedAt:   f.CreatedAt,
			UpdatedAt:   f.UpdatedAt,
		})
	}

	for _, f := range result.Folders {
		reply.Folders = append(reply.Folders, &pb.GatewayFolderMetadataReply{
			Id:        f.ID,
			Name:      f.Name,
			ParentId:  f.ParentID,
			OwnerId:   f.OwnerID,
			CreatedAt: f.CreatedAt,
			UpdatedAt: f.UpdatedAt,
		})
	}

	return reply, nil
}
