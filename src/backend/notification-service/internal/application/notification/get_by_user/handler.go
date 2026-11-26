package getbyuser

import (
	"context"

	"notification-service/internal/application/notification"
	"notification-service/internal/domain/model"

	"github.com/gocql/gocql"
)

type Handler struct {
	repo notification.Repository
}

func NewHandler(repo notification.Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

type Query struct {
	UserID gocql.UUID
}

type ResponseItem struct {
	UserID         gocql.UUID
	NotificationID gocql.UUID
	EventType      model.NotificationEventType
	Payload        model.Payload
}

type Response struct {
	Items []*ResponseItem
}

func (h *Handler) Handle(ctx context.Context, query Query) (*Response, error) {
	res, err := h.repo.GetNotificationsByUserID(ctx, query.UserID)
	if err != nil {
		return nil, err
	}

	var items []*ResponseItem
	for _, resItem := range res {
		items = append(items, &ResponseItem{
			UserID:         resItem.UserID,
			NotificationID: resItem.NotificationID,
			EventType:      resItem.EventType,
			Payload:        resItem.Payload,
		})
	}

	return &Response{
		Items: items,
	}, nil
}
