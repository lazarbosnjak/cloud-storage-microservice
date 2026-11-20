package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"file-service/internal/application/file"
	"file-service/internal/application/folder"
	"file-service/internal/application/ping"
	"file-service/internal/domain/models"

	pb "proto/file"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func newNetworkTestClient(
	t *testing.T,
) (pb.FileClient, *fakeFolderRepository, *fakeGRPCFileRepository) {
	t.Helper()

	folderRepo := newFakeFolderRepository()
	fileRepo := newFakeGRPCFileRepository()

	server := NewServer(
		ping.NewHandler(0),
		folder.NewCreateHandler(folderRepo),
		file.NewRenameHandler(fileRepo),
		folder.NewRenameHandler(folderRepo),
		folder.NewListContentsHandler(folderRepo, fileRepo),
		file.NewGetHandler(fileRepo),
		folder.NewGetHandler(folderRepo),
	)

	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()

	pb.RegisterFileServer(grpcServer, server)

	go func() {
		_ = grpcServer.Serve(listener)
	}()

	t.Cleanup(func() {
		grpcServer.Stop()
		_ = listener.Close()
	})

	dialer := func(
		ctx context.Context,
		_ string,
	) (net.Conn, error) {
		return listener.DialContext(ctx)
	}

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
	})

	return pb.NewFileClient(conn), folderRepo, fileRepo
}

func TestGRPCNetworkMetadataWorkflow(t *testing.T) {
	client, _, fileRepo := newNetworkTestClient(t)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	created, err := client.CreateFolder(
		ctx,
		&pb.CreateFolderRequest{
			Name:    "Documents",
			OwnerId: "user-1",
		},
	)
	if err != nil {
		t.Fatalf("CreateFolder RPC failed: %v", err)
	}

	if created.GetId() == "" {
		t.Fatal("expected created folder ID")
	}

	folderMetadata, err := client.GetFolder(
		ctx,
		&pb.GetFolderRequest{
			Id:      created.GetId(),
			OwnerId: "user-1",
		},
	)
	if err != nil {
		t.Fatalf("GetFolder RPC failed: %v", err)
	}

	if folderMetadata.GetName() != "Documents" {
		t.Fatal("unexpected folder name")
	}

	_, err = client.RenameFolder(
		ctx,
		&pb.RenameFolderRequest{
			Id:      created.GetId(),
			OwnerId: "user-1",
			Name:    "RenamedDocuments",
		},
	)
	if err != nil {
		t.Fatalf("RenameFolder RPC failed: %v", err)
	}

	fileRepo.files["file-1"] = &models.FileMetadata{
		ID:       "file-1",
		Name:     "old.txt",
		ParentID: created.GetId(),
		OwnerID:  "user-1",
		Size:     512,
	}

	_, err = client.RenameFile(
		ctx,
		&pb.RenameFileRequest{
			Id:      "file-1",
			OwnerId: "user-1",
			Name:    "new.txt",
		},
	)
	if err != nil {
		t.Fatalf("RenameFile RPC failed: %v", err)
	}

	fileMetadata, err := client.GetFile(
		ctx,
		&pb.GetFileRequest{
			Id:      "file-1",
			OwnerId: "user-1",
		},
	)
	if err != nil {
		t.Fatalf("GetFile RPC failed: %v", err)
	}

	if fileMetadata.GetName() != "new.txt" {
		t.Fatalf("unexpected file name: %s", fileMetadata.GetName())
	}

	contents, err := client.ListFolderContents(
		ctx,
		&pb.ListFolderContentsRequest{
			FolderId: created.GetId(),
			OwnerId:  "user-1",
		},
	)
	if err != nil {
		t.Fatalf("ListFolderContents RPC failed: %v", err)
	}

	if len(contents.GetFiles()) != 1 {
		t.Fatalf("expected 1 file, got %d", len(contents.GetFiles()))
	}

	if contents.GetFiles()[0].GetName() != "new.txt" {
		t.Fatal("unexpected listed file name")
	}

	updatedFolder, err := client.GetFolder(
		ctx,
		&pb.GetFolderRequest{
			Id:      created.GetId(),
			OwnerId: "user-1",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if updatedFolder.GetName() != "RenamedDocuments" {
		t.Fatal("folder rename was not persisted")
	}
}

func TestGRPCNetworkOwnerIsolation(t *testing.T) {
	client, _, _ := newNetworkTestClient(t)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	created, err := client.CreateFolder(
		ctx,
		&pb.CreateFolderRequest{
			Name:    "Private",
			OwnerId: "user-1",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetFolder(
		ctx,
		&pb.GetFolderRequest{
			Id:      created.GetId(),
			OwnerId: "user-2",
		},
	)

	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}

	_, err = client.RenameFolder(
		ctx,
		&pb.RenameFolderRequest{
			Id:      created.GetId(),
			OwnerId: "user-2",
			Name:    "Unauthorized",
		},
	)

	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}
