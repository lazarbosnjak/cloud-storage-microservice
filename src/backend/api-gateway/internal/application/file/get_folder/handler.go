package get_folder

import (
	"context"

	"api-gateway/internal/application/file"
)

type Query struct {
	ID      string
	OwnerID string
}

type Handler struct {
	fileClient file.Client
}

func NewHandler(fileClient file.Client) *Handler {
	return &Handler{
		fileClient: fileClient,
	}
}

func (h *Handler) Handle(
	ctx context.Context,
	query Query,
) (*file.FolderMetadata, error) {
	return h.fileClient.GetFolder(ctx, query.ID, query.OwnerID)
}
