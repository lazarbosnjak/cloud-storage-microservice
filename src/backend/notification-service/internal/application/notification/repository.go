package notification

import (
	"context"

	"notification-service/internal/domain/model"

	"github.com/gocql/gocql"
)

type Repository interface {
	GetNotificationsByUserID(ctx context.Context, userID gocql.UUID) ([]model.NotificationByUserID, error)
}
