package folder

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"file-service/internal/domain/models"
	"file-service/internal/domain/repository"
)

type ListContentsHandler struct {
	folderRepository repository.FolderRepository
	fileRepository   repository.FileRepository
}

func NewListContentsHandler(
	folderRepo repository.FolderRepository,
	fileRepo repository.FileRepository,
) *ListContentsHandler {
	return &ListContentsHandler{
		folderRepository: folderRepo,
		fileRepository:   fileRepo,
	}
}

type ListContentsQuery struct {
	FolderID string
	OwnerID  string
}

type ListContentsResult struct {
	Folders []*models.FolderMetadata
	Files   []*models.FileMetadata
}

func (h *ListContentsHandler) Handle(
	ctx context.Context,
	query ListContentsQuery,
) (*ListContentsResult, error) {
	folderID := strings.TrimSpace(query.FolderID)
	ownerID := strings.TrimSpace(query.OwnerID)

	if ownerID == "" {
		return nil, ErrInvalidOwner
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Prazan FolderID predstavlja root direktorijum.
	if folderID != "" {
		parent, err := h.folderRepository.GetByID(ctx, folderID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, repository.ErrNotFound
			}

			return nil, fmt.Errorf("get parent folder: %w", err)
		}

		if parent.OwnerID != ownerID {
			return nil, repository.ErrNotFound
		}
	}

	folders, err := h.folderRepository.ListByParentID(
		ctx,
		folderID,
		ownerID,
	)
	if err != nil {
		slog.Error("failed to list folders", "err", err)
		return nil, fmt.Errorf("list folders: %w", err)
	}

	files, err := h.fileRepository.ListByParentID(
		ctx,
		folderID,
		ownerID,
	)
	if err != nil {
		slog.Error("failed to list files", "err", err)
		return nil, fmt.Errorf("list files: %w", err)
	}

	return &ListContentsResult{
		Folders: folders,
		Files:   files,
	}, nil
}
