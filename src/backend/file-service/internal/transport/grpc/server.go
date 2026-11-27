package grpc

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"file-service/internal/application/file"
	"file-service/internal/application/folder"
	"file-service/internal/application/ping"
	"file-service/internal/domain/models"
	"file-service/internal/domain/quota"
	"file-service/internal/domain/repository"

	pb "proto/file"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedFileServer

	pingHandler         *ping.Handler
	createFileHandler   *file.UploadHandler
	createFolderHandler *folder.CreateHandler
	renameFileHandler   *file.RenameHandler
	renameFolderHandler *folder.RenameHandler
	listContentsHandler *folder.ListContentsHandler
	getFileHandler      *file.GetHandler
	getFolderHandler    *folder.GetHandler
}

func NewServer(
	pingHandler *ping.Handler,
	createFileHandler *file.UploadHandler,
	createFolderHandler *folder.CreateHandler,
	renameFileHandler *file.RenameHandler,
	renameFolderHandler *folder.RenameHandler,
	listContentsHandler *folder.ListContentsHandler,
	getFileHandler *file.GetHandler,
	getFolderHandler *folder.GetHandler,
) *Server {
	return &Server{
		pingHandler:         pingHandler,
		createFileHandler:   createFileHandler,
		createFolderHandler: createFolderHandler,
		renameFileHandler:   renameFileHandler,
		renameFolderHandler: renameFolderHandler,
		listContentsHandler: listContentsHandler,
		getFileHandler:      getFileHandler,
		getFolderHandler:    getFolderHandler,
	}
}

func mapError(err error) error {
	switch {
	case errors.Is(err, quota.ErrInsufficientStorage):
		slog.Warn("file upload rejected: insufficient storage")
		return status.Error(codes.ResourceExhausted, "insufficient storage")

	case errors.Is(err, quota.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")

	case errors.Is(err, quota.ErrInvalidStorageUsage):
		slog.Warn("invalid storage usage")
		return status.Error(codes.FailedPrecondition, "invalid storage usage")

	case errors.Is(err, quota.ErrStorageUnavailable):
		slog.Error("user storage unavailable", "err", err)
		return status.Error(codes.Unavailable, "user storage unavailable")

	case errors.Is(err, repository.ErrFileAlreadyExists):
		return status.Error(codes.AlreadyExists, "file with the same name already exists")

	case errors.Is(err, repository.ErrVersionConflict):
		slog.Warn("file version conflict")
		return status.Error(codes.Aborted, "file version conflict")

	case errors.Is(err, repository.ErrNotFound),
		errors.Is(err, folder.ErrInvalidParent),
		errors.Is(err, file.ErrInvalidParent):
		return status.Error(codes.NotFound, "resource not found")

	case errors.Is(err, file.ErrInvalidFileID),
		errors.Is(err, folder.ErrInvalidFolderID),
		errors.Is(err, file.ErrInvalidName),
		errors.Is(err, folder.ErrInvalidName),
		errors.Is(err, file.ErrInvalidOwner),
		errors.Is(err, folder.ErrInvalidOwner),
		errors.Is(err, file.ErrInvalidSize),
		errors.Is(err, file.ErrInvalidContentType),
		errors.Is(err, file.ErrInvalidStorageKey):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())

	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())

	default:
		slog.Error("file service RPC failed", "err", err)
		return status.Error(codes.Internal, "internal server error")
	}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.UTC().Format(time.RFC3339Nano)
}

func toFileMetadataReply(
	value *models.FileMetadata,
) *pb.FileMetadataReply {
	return &pb.FileMetadataReply{
		Id:          value.ID,
		Name:        value.Name,
		Size:        value.Size,
		ContentType: value.ContentType,
		ParentId:    value.ParentID,
		OwnerId:     value.OwnerID,
		CreatedAt:   formatTime(value.CreatedAt),
		UpdatedAt:   formatTime(value.UpdatedAt),
	}
}

func toFolderMetadataReply(
	value *models.FolderMetadata,
) *pb.FolderMetadataReply {
	return &pb.FolderMetadataReply{
		Id:        value.ID,
		Name:      value.Name,
		ParentId:  value.ParentID,
		OwnerId:   value.OwnerID,
		CreatedAt: formatTime(value.CreatedAt),
		UpdatedAt: formatTime(value.UpdatedAt),
	}
}

func (s *Server) Ping(
	ctx context.Context,
	_ *pb.PingRequest,
) (*pb.PingReply, error) {
	message, err := s.pingHandler.Handle(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.PingReply{Message: message}, nil
}

func (s *Server) CreateFile(
	ctx context.Context,
	req *pb.CreateFileRequest,
) (*pb.CreateFileReply, error) {
	if req == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"request cannot be nil",
		)
	}

	result, err := s.createFileHandler.Handle(
		ctx,
		file.UploadCommand{
			Name:        req.GetName(),
			Size:        req.GetSize(),
			ContentType: req.GetContentType(),
			ParentID:    req.GetParentId(),
			OwnerID:     req.GetOwnerId(),
		},
	)
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.CreateFileReply{
		Id:          result.ID,
		Name:        result.Name,
		Size:        result.Size,
		ContentType: result.ContentType,
		StorageKey:  result.StorageKey,
		ParentId:    result.ParentID,
		OwnerId:     result.OwnerID,
		CreatedAt:   formatTime(result.CreatedAt),
		UpdatedAt:   formatTime(result.UpdatedAt),
	}, nil
}

func (s *Server) CreateFolder(
	ctx context.Context,
	req *pb.CreateFolderRequest,
) (*pb.CreateFolderReply, error) {
	if req == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"request cannot be nil",
		)
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
		return nil, mapError(err)
	}

	return &pb.CreateFolderReply{
		Id:        result.ID,
		Name:      result.Name,
		ParentId:  result.ParentID,
		OwnerId:   result.OwnerID,
		CreatedAt: formatTime(result.CreatedAt),
	}, nil
}

func (s *Server) RenameFile(
	ctx context.Context,
	req *pb.RenameFileRequest,
) (*pb.RenameFileReply, error) {
	if req == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"request cannot be nil",
		)
	}

	err := s.renameFileHandler.Handle(
		ctx,
		file.RenameCommand{
			ID:      req.GetId(),
			OwnerID: req.GetOwnerId(),
			Name:    req.GetName(),
		},
	)
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.RenameFileReply{}, nil
}

func (s *Server) RenameFolder(
	ctx context.Context,
	req *pb.RenameFolderRequest,
) (*pb.RenameFolderReply, error) {
	if req == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"request cannot be nil",
		)
	}

	err := s.renameFolderHandler.Handle(
		ctx,
		folder.RenameCommand{
			ID:      req.GetId(),
			OwnerID: req.GetOwnerId(),
			Name:    req.GetName(),
		},
	)
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.RenameFolderReply{}, nil
}

func (s *Server) ListFolderContents(
	ctx context.Context,
	req *pb.ListFolderContentsRequest,
) (*pb.ListFolderContentsReply, error) {
	if req == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"request cannot be nil",
		)
	}

	result, err := s.listContentsHandler.Handle(
		ctx,
		folder.ListContentsQuery{
			FolderID: req.GetFolderId(),
			OwnerID:  req.GetOwnerId(),
		},
	)
	if err != nil {
		return nil, mapError(err)
	}

	reply := &pb.ListFolderContentsReply{
		Folders: make([]*pb.FolderMetadataReply, 0, len(result.Folders)),
		Files:   make([]*pb.FileMetadataReply, 0, len(result.Files)),
	}

	for _, folderMetadata := range result.Folders {
		reply.Folders = append(
			reply.Folders,
			toFolderMetadataReply(folderMetadata),
		)
	}

	for _, fileMetadata := range result.Files {
		reply.Files = append(
			reply.Files,
			toFileMetadataReply(fileMetadata),
		)
	}

	return reply, nil
}

func (s *Server) GetFile(
	ctx context.Context,
	req *pb.GetFileRequest,
) (*pb.FileMetadataReply, error) {
	if req == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"request cannot be nil",
		)
	}

	result, err := s.getFileHandler.Handle(
		ctx,
		file.GetQuery{
			ID:      req.GetId(),
			OwnerID: req.GetOwnerId(),
		},
	)
	if err != nil {
		return nil, mapError(err)
	}

	return toFileMetadataReply(result), nil
}

func (s *Server) GetFolder(
	ctx context.Context,
	req *pb.GetFolderRequest,
) (*pb.FolderMetadataReply, error) {
	if req == nil {
		return nil, status.Error(
			codes.InvalidArgument,
			"request cannot be nil",
		)
	}

	result, err := s.getFolderHandler.Handle(
		ctx,
		folder.GetQuery{
			ID:      req.GetId(),
			OwnerID: req.GetOwnerId(),
		},
	)
	if err != nil {
		return nil, mapError(err)
	}

	return toFolderMetadataReply(result), nil
}
