package file

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

func (h *Handler) GetFile(
	ctx context.Context,
	id string,
	ownerID string,
) (*appfile.FileMetadata, error) {
	return h.client.GetFile(ctx, id, ownerID)
}

func (h *Handler) CreateFile(
	ctx context.Context,
	input appfile.CreateFileInput,
) (*appfile.CreatedFile, error) {
	return h.client.CreateFile(ctx, input)
}

func (h *Handler) RenameFile(
	ctx context.Context,
	id, ownerID, name string,
) error {
	return h.client.RenameFile(ctx, id, ownerID, name)
}
