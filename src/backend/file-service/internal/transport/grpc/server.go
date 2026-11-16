package grpc

import (
	"context"
	"errors"

	"file-service/internal/application/folder"
	"file-service/internal/application/ping"

	pb "proto/file"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedFileServer

	pingHandler         *ping.Handler
	createFolderHandler *folder.CreateHandler
}

func NewServer(
	pingHandler *ping.Handler,
	createFolderHandler *folder.CreateHandler,
) *Server {
	return &Server{
		pingHandler:         pingHandler,
		createFolderHandler: createFolderHandler,
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

func (s *Server) CreateFolder(
	ctx context.Context,
	req *pb.CreateFolderRequest,
) (*pb.CreateFolderReply, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	result, err := s.createFolderHandler.Handle(
		ctx,
		folder.CreateCommand{
			Name:     req.GetName(),
			ParentID: req.GetParentId(),
			OwnerID:  req.GetOwnerId(),
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, folder.ErrInvalidName),
			errors.Is(err, folder.ErrInvalidOwner):
			return nil, status.Error(codes.InvalidArgument, err.Error())

		case errors.Is(err, folder.ErrInvalidParent):
			return nil, status.Error(codes.NotFound, err.Error())

		case errors.Is(err, context.Canceled):
			return nil, status.Error(codes.Canceled, err.Error())

		case errors.Is(err, context.DeadlineExceeded):
			return nil, status.Error(codes.DeadlineExceeded, err.Error())

		default:
			return nil, status.Error(codes.Internal, "failed to create folder")
		}
	}

	return &pb.CreateFolderReply{
		Id:        result.ID,
		Name:      result.Name,
		ParentId:  result.ParentID,
		OwnerId:   result.OwnerID,
		CreatedAt: result.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
	}, nil
}
