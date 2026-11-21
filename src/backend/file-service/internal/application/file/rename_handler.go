package file

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"file-service/internal/domain/repository"
)

var (
	ErrInvalidFileID = errors.New("file ID cannot be empty")
	ErrInvalidOwner  = errors.New("file owner cannot be empty")
	ErrInvalidName   = errors.New("file name cannot be empty")
)

type RenameHandler struct {
	repository repository.FileRepository
}

func NewRenameHandler(repo repository.FileRepository) *RenameHandler {
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
		slog.Warn("file rename rejected: empty file ID")
		return ErrInvalidFileID
	}

	if ownerID == "" {
		slog.Warn("file rename rejected: empty owner ID")
		return ErrInvalidOwner
	}

	if name == "" {
		slog.Warn("file rename rejected: empty name")
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

		slog.Error("failed to rename file", "err", err)
		return fmt.Errorf("rename file: %w", err)
	}

	slog.Info("file renamed successfully", "file_id", id)
	return nil
}
