package grpc

import (
	"context"
	"errors"

	pb "proto/user"
	storageapp "user-service/internal/application/user/storage"
	"user-service/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GetStorage(
	ctx context.Context,
	in *pb.GetStorageRequest,
) (*pb.StorageReply, error) {
	storage, err := s.StorageHandler.GetStorage(
		ctx,
		in.GetUserId(),
	)

	if err != nil {
		return nil, storageError(err)
	}

	return toStorageReply(storage), nil
}

func (s *Server) AdjustStorageUsage(
	ctx context.Context,
	in *pb.AdjustStorageUsageRequest,
) (*pb.StorageReply, error) {
	storage, err := s.StorageHandler.AdjustStorageUsage(
		ctx,
		in.GetUserId(),
		in.GetDelta(),
	)

	if err != nil {
		return nil, storageError(err)
	}

	return toStorageReply(storage), nil
}

func toStorageReply(storage domain.StorageInfo) *pb.StorageReply {
	available := storage.Quota - storage.Used

	if available < 0 {
		available = 0
	}

	return &pb.StorageReply{
		Quota:     storage.Quota,
		Used:      storage.Used,
		Available: available,
	}
}

func storageError(err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")

	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "request deadline exceeded")

	case errors.Is(err, storageapp.ErrInvalidUserID):
		return status.Error(codes.InvalidArgument, "invalid user ID")

	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")

	case errors.Is(err, domain.ErrInsufficientStorage):
		return status.Error(codes.ResourceExhausted, "insufficient storage")

	case errors.Is(err, domain.ErrInvalidStorageUsage):
		return status.Error(codes.FailedPrecondition, "invalid storage usage")

	case errors.Is(err, domain.ErrStorageNotInitialized):
		return status.Error(codes.FailedPrecondition, "storage not initialized")

	default:
		return status.Error(codes.Internal, "storage operation failed")
	}
}
