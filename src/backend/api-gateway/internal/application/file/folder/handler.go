package folder

import (
	"context"

	appfile "api-gateway/internal/application/file"
)

type Handler struct {
	client appfile.Client
}

func NewHandler(client appfile.Client) *Handler {
	return &Handler{
		client: client,
	}
}

func (h *Handler) GetFolder(
	ctx context.Context,
	id string,
	ownerID string,
) (*appfile.FolderMetadata, error) {
	return h.client.GetFolder(ctx, id, ownerID)
}

func (h *Handler) CreateFolder(
	ctx context.Context,
	input appfile.CreateFolderInput,
) (*appfile.CreatedFolder, error) {
	return h.client.CreateFolder(ctx, input)
}

func (h *Handler) RenameFolder(
	ctx context.Context,
	id, ownerID, name string,
) error {
	return h.client.RenameFolder(ctx, id, ownerID, name)
}

func (h *Handler) ListFolderContents(
	ctx context.Context,
	folderID, ownerID string,
) (*appfile.FolderContents, error) {
	return h.client.ListFolderContents(ctx, folderID, ownerID)
}
