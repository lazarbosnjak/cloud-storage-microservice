package grpc

import (
	"context"
	"testing"
	"time"

	"file-service/internal/application/file"
	"file-service/internal/application/folder"
	"file-service/internal/domain/models"

	pb "proto/file"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGRPCCreateFileSuccess(t *testing.T) {
	fileRepo := newFakeGRPCFileRepository()
	folderRepo := newFakeFolderRepository()

	folderRepo.folders["folder-1"] = &models.FolderMetadata{
		ID:      "folder-1",
		Name:    "Documents",
		OwnerID: "user-1",
	}

	server := &Server{
		createFileHandler: file.NewCreateHandler(
			fileRepo,
			folderRepo,
		),
	}

	reply, err := server.CreateFile(
		context.Background(),
		&pb.CreateFileRequest{
			Name:        "report.pdf",
			Size:        2048,
			ContentType: "application/pdf",
			StorageKey:  "files/report.pdf",
			ParentId:    "folder-1",
			OwnerId:     "user-1",
		},
	)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	if reply.GetId() == "" {
		t.Fatal("expected generated file ID")
	}

	if reply.GetName() != "report.pdf" {
		t.Fatalf("unexpected name: %s", reply.GetName())
	}

	if reply.GetSize() != 2048 {
		t.Fatalf("unexpected size: %d", reply.GetSize())
	}

	if reply.GetParentId() != "folder-1" {
		t.Fatal("incorrect parent folder")
	}

	if reply.GetOwnerId() != "user-1" {
		t.Fatal("incorrect owner")
	}

	if reply.GetCreatedAt() == "" || reply.GetUpdatedAt() == "" {
		t.Fatal("expected timestamps")
	}

	saved := fileRepo.files[reply.GetId()]
	if saved == nil {
		t.Fatal("file metadata was not saved")
	}

	if saved.StorageKey != "files/report.pdf" {
		t.Fatal("incorrect storage key")
	}
}

func TestGRPCCreateFileInvalidRequests(t *testing.T) {
	tests := []struct {
		name     string
		request  *pb.CreateFileRequest
		wantCode codes.Code
	}{
		{
			name:     "nil request",
			request:  nil,
			wantCode: codes.InvalidArgument,
		},
		{
			name: "empty name",
			request: &pb.CreateFileRequest{
				OwnerId:     "user-1",
				Size:        100,
				ContentType: "text/plain",
				StorageKey:  "files/test.txt",
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "empty owner",
			request: &pb.CreateFileRequest{
				Name:        "test.txt",
				Size:        100,
				ContentType: "text/plain",
				StorageKey:  "files/test.txt",
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "zero size",
			request: &pb.CreateFileRequest{
				Name:        "test.txt",
				OwnerId:     "user-1",
				ContentType: "text/plain",
				StorageKey:  "files/test.txt",
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "missing parent",
			request: &pb.CreateFileRequest{
				Name:        "test.txt",
				OwnerId:     "user-1",
				ParentId:    "missing-folder",
				Size:        100,
				ContentType: "text/plain",
				StorageKey:  "files/test.txt",
			},
			wantCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileRepo := newFakeGRPCFileRepository()
			folderRepo := newFakeFolderRepository()

			server := &Server{
				createFileHandler: file.NewCreateHandler(
					fileRepo,
					folderRepo,
				),
			}

			_, err := server.CreateFile(
				context.Background(),
				tt.request,
			)

			if status.Code(err) != tt.wantCode {
				t.Fatalf(
					"expected %v, got %v: %v",
					tt.wantCode,
					status.Code(err),
					err,
				)
			}

			if len(fileRepo.files) != 0 {
				t.Fatal("invalid file must not be saved")
			}
		})
	}
}

func TestGRPCCreateFileDifferentOwner(t *testing.T) {
	fileRepo := newFakeGRPCFileRepository()
	folderRepo := newFakeFolderRepository()

	folderRepo.folders["private-folder"] = &models.FolderMetadata{
		ID:      "private-folder",
		OwnerID: "user-2",
	}

	server := &Server{
		createFileHandler: file.NewCreateHandler(
			fileRepo,
			folderRepo,
		),
	}

	_, err := server.CreateFile(
		context.Background(),
		&pb.CreateFileRequest{
			Name:        "private.txt",
			Size:        100,
			ContentType: "text/plain",
			StorageKey:  "files/private.txt",
			ParentId:    "private-folder",
			OwnerId:     "user-1",
		},
	)

	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}

	if len(fileRepo.files) != 0 {
		t.Fatal("file must not be saved in another user's folder")
	}
}

func TestGRPCCreateFileCanceledContext(t *testing.T) {
	fileRepo := newFakeGRPCFileRepository()
	folderRepo := newFakeFolderRepository()

	server := &Server{
		createFileHandler: file.NewCreateHandler(
			fileRepo,
			folderRepo,
		),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := server.CreateFile(
		ctx,
		&pb.CreateFileRequest{
			Name:        "test.txt",
			Size:        100,
			ContentType: "text/plain",
			StorageKey:  "files/test.txt",
			OwnerId:     "user-1",
		},
	)

	if status.Code(err) != codes.Canceled {
		t.Fatalf("expected Canceled, got %v", err)
	}
}

func TestGRPCCreateFileTimestampFormat(t *testing.T) {
	fileRepo := newFakeGRPCFileRepository()
	folderRepo := newFakeFolderRepository()

	server := &Server{
		createFileHandler: file.NewCreateHandler(
			fileRepo,
			folderRepo,
		),
	}

	reply, err := server.CreateFile(
		context.Background(),
		&pb.CreateFileRequest{
			Name:        "notes.txt",
			Size:        100,
			ContentType: "text/plain",
			StorageKey:  "files/notes.txt",
			OwnerId:     "user-1",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = time.Parse(time.RFC3339Nano, reply.GetCreatedAt())
	if err != nil {
		t.Fatalf("invalid created_at timestamp: %v", err)
	}

	_, err = time.Parse(time.RFC3339Nano, reply.GetUpdatedAt())
	if err != nil {
		t.Fatalf("invalid updated_at timestamp: %v", err)
	}
}

// Ensure the existing folder handler remains available.
var _ = folder.CreateCommand{}
