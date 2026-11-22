package folder

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"file-service/internal/domain/repository"
)

var ErrInvalidFolderID = errors.New("folder ID cannot be empty")

type RenameHandler struct {
	repository repository.FolderRepository
}

func NewRenameHandler(repo repository.FolderRepository) *RenameHandler {
	return &RenameHandler{repository: repo}
}

type RenameCommand struct {
	ID      string
	OwnerID string
	Name    string
}

func (h *RenameHandler) Handle(ctx context.Context, cmd RenameCommand) error {
	id := strings.TrimSpace(cmd.ID)
	ownerID := strings.TrimSpace(cmd.OwnerID)
	name := strings.TrimSpace(cmd.Name)

	if id == "" {
		slog.Warn("folder rename rejected: empty folder ID")
		return ErrInvalidFolderID
	}

	if ownerID == "" {
		slog.Warn("folder rename rejected: empty owner ID")
		return ErrInvalidOwner
	}

	if name == "" {
		slog.Warn("folder rename rejected: empty name")
		return ErrInvalidName
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if err := h.repository.Rename(
		ctx,
		id,
		ownerID,
		name,
		time.Now().UTC(),
	); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return err
		}

		slog.Error("failed to rename folder", "err", err)
		return fmt.Errorf("rename folder: %w", err)
	}

	slog.Info("folder renamed successfully", "folder_id", id)
	return nil
}
