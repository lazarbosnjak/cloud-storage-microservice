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
