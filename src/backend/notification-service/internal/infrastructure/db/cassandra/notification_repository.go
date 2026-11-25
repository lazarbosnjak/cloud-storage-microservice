package cassandra

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"notification-service/internal/domain/model"
	"notification-service/internal/startup/config"

	"github.com/gocql/gocql"
)

const keyspace = "notification"

type NotificationsRepository struct {
	session *gocql.Session
}

func NewNotificationsRepository(cfg *config.Config) (*NotificationsRepository, error) {
	dbAddr := cfg.Cassandra.Addr

	// Connect to default keyspace
	cluster := gocql.NewCluster(dbAddr)
	cluster.Keyspace = "system"
	session, err := cluster.CreateSession()
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	// Create 'notification' keyspace
	err = session.Query(
		fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s
					WITH replication = {
						'class' : 'SimpleStrategy',
						'replication_factor' : %d
					}`, keyspace, 1)).Exec()
	if err != nil {
		slog.Error(err.Error())
	}
	session.Close()

	// Connect to student keyspace
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.One
	session, err = cluster.CreateSession()
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	notifRepo := &NotificationsRepository{
		session: session,
	}

	if err := notifRepo.createTables(); err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	return notifRepo, nil
}

func (r *NotificationsRepository) createTables() error {
	err := r.session.Query(
		fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS %s (
			user_id uuid,
			notification_id timeuuid,

			event_type text,

			payload_schema_version int,
			payload text,

			actor_id uuid,

			PRIMARY KEY ((user_id), notification_id)) 
			WITH CLUSTERING ORDER BY (notification_id DESC)`,
			"notifications_by_user")).Exec()
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func (r *NotificationsRepository) GetNotificationsByUserId(
	ctx context.Context,
	userID string,
) ([]*model.NotificationByUserID, error) {
	query := `
	SELECT user_id, notification_id, event_type, payload_schema_version, payload, actor_id
	FROM notifications_by_user
	WHERE user_id = ?
	`

	scanner := r.session.Query(query, userID).Iter().Scanner()

	var notifications []*model.NotificationByUserID
	for scanner.Next() {
		var n model.NotificationByUserID
		err := scanner.Scan(
			n.UserID, n.NotificationID, &n.EventType,
			&n.PayloadSchemaVersion, &n.Payload, &n.ActorID)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, &n)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *NotificationsRepository) InsertNotificationsByUserId(
	n *model.NotificationByUserID,
) error {
	n.NotificationID = gocql.TimeUUID()

	query := `
	INSERT INTO notifications_by_user (
			user_id,
			notification_id,
			event_type,
			payload_schema_version,
			payload,
			actor_id)
	VALUES (?, ?, ?, ?, ?, ?)`

	payload, err := json.Marshal(n.Payload)
	if err != nil {
		return err
	}

	err = r.session.Query(query,
		n.UserID, n.NotificationID, n.EventType,
		n.PayloadSchemaVersion, payload, n.ActorID).Exec()
	if err != nil {
		slog.Error("error inserting notification by user id", "err", err)
		return err
	}

	return nil
}

// TODO:
// - Create mock data
// - Fetch mock data
// - Create 2 more tables
// - Create RPC
