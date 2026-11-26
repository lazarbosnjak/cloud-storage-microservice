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
		createFileHandler: file.NewUploadHandler(
			fileRepo,
			folderRepo,
			newFakeQuotaClient(),
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

	if saved.StorageKey != "placeholder/"+reply.GetId() {
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
				createFileHandler: file.NewUploadHandler(
					fileRepo,
					folderRepo,
					newFakeQuotaClient(),
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
		createFileHandler: file.NewUploadHandler(
			fileRepo,
			folderRepo,
			newFakeQuotaClient(),
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
		createFileHandler: file.NewUploadHandler(
			fileRepo,
			folderRepo,
			newFakeQuotaClient(),
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
		createFileHandler: file.NewUploadHandler(
			fileRepo,
			folderRepo,
			newFakeQuotaClient(),
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

func TestGRPCCreateFileVersioning(t *testing.T) {
	fileRepo := newFakeGRPCFileRepository()
	folderRepo := newFakeFolderRepository()
	quotaClient := newFakeQuotaClient()

	server := &Server{
		createFileHandler: file.NewUploadHandler(
			fileRepo,
			folderRepo,
			quotaClient,
		),
	}

	ctx := context.Background()

	request := &pb.CreateFileRequest{
		Name:        "report.pdf",
		Size:        1000,
		ContentType: "application/pdf",
		OwnerId:     "user-1",
	}

	first, err := server.CreateFile(ctx, request)
	if err != nil {
		t.Fatal(err)
	}

	request.Size = 1500

	second, err := server.CreateFile(ctx, request)
	if err != nil {
		t.Fatal(err)
	}

	if first.GetId() != second.GetId() {
		t.Fatal("file ID must remain unchanged")
	}

	if first.GetCreatedAt() != second.GetCreatedAt() {
		t.Fatal("creation timestamp must remain unchanged")
	}

	if first.GetUpdatedAt() == second.GetUpdatedAt() {
		t.Fatal("expected updated timestamp")
	}

	if second.GetSize() != 1500 {
		t.Fatal("incorrect updated file size")
	}

	if len(fileRepo.files) != 1 {
		t.Fatal("expected exactly one file")
	}

	storage, err := quotaClient.GetStorage(ctx, "user-1")
	if err != nil {
		t.Fatal(err)
	}

	if storage.Used != 1500 {
		t.Fatalf("expected usage 1500, got %d", storage.Used)
	}
}

func TestGRPCCreateFileQuotaExceeded(t *testing.T) {
	fileRepo := newFakeGRPCFileRepository()
	folderRepo := newFakeFolderRepository()
	quotaClient := newFakeQuotaClient()

	quotaClient.limit = 1000

	server := &Server{
		createFileHandler: file.NewUploadHandler(
			fileRepo,
			folderRepo,
			quotaClient,
		),
	}

	ctx := context.Background()

	request := &pb.CreateFileRequest{
		Name:        "report.pdf",
		Size:        800,
		ContentType: "application/pdf",
		OwnerId:     "user-1",
	}

	first, err := server.CreateFile(ctx, request)
	if err != nil {
		t.Fatal(err)
	}

	request.Size = 1200

	_, err = server.CreateFile(ctx, request)

	if status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("expected ResourceExhausted, got %v", err)
	}

	if len(fileRepo.files) != 1 {
		t.Fatal("unexpected number of files")
	}

	if fileRepo.files[first.GetId()].Size != 800 {
		t.Fatal("original file must remain unchanged")
	}

	storage, err := quotaClient.GetStorage(ctx, "user-1")
	if err != nil {
		t.Fatal(err)
	}

	if storage.Used != 800 {
		t.Fatalf("expected usage 800, got %d", storage.Used)
	}
}
