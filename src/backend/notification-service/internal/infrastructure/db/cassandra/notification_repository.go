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

	if err := notifRepo.insertMockData(context.Background()); err != nil {
		slog.Error("error initiating mock data", "err", err)
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

			PRIMARY KEY ((user_id), notification_id)) 
			WITH CLUSTERING ORDER BY (notification_id DESC)`,
			"notifications_by_user")).Exec()
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func (r *NotificationsRepository) GetNotificationsByUserID(
	ctx context.Context,
	userID gocql.UUID,
) ([]model.NotificationByUserID, error) {
	query := `
	SELECT notification_id, event_type, payload_schema_version, payload
	FROM notifications_by_user
	WHERE user_id = ?
	`

	iter := r.session.Query(query, userID).WithContext(ctx).Iter()

	var (
		out []model.NotificationByUserID
		id  gocql.UUID
		typ model.NotificationEventType
		ver int
		raw string
	)
	for iter.Scan(&id, &typ, &ver, &raw) {
		p, err := decodePayload(typ, raw)
		if err != nil {
			slog.Error("error decoding payload", "err", err)
			continue
		}
		out = append(out, model.NotificationByUserID{
			UserID:         userID,
			NotificationID: id,
			EventType:      typ,
			Payload:        p,
		})
	}
	if err := iter.Close(); err != nil {
		slog.Error("error fetching notifications by user id", "err", err)
		return nil, err
	}

	return out, nil
}

func (r *NotificationsRepository) InsertNotificationsByUserID(
	ctx context.Context,
	userID gocql.UUID,
	p model.Payload,
) (gocql.UUID, error) {
	notificationID := gocql.TimeUUID()

	query := `
	INSERT INTO notifications_by_user (
			user_id,
			notification_id,
			event_type,
			payload_schema_version,
			payload)
	VALUES (?, ?, ?, ?, ?)`

	payload, err := json.Marshal(p)
	if err != nil {
		return gocql.UUID{}, err
	}

	err = r.session.Query(query,
		userID, notificationID,
		p.EventType(), p.SchemaVersion(), string(payload)).WithContext(ctx).Exec()
	if err != nil {
		slog.Error("error inserting notification by user id", "err", err)
		return gocql.UUID{}, err
	}

	return notificationID, nil
}

func (r *NotificationsRepository) insertMockData(ctx context.Context) error {
	userID, err := gocql.ParseUUID("01a2142e-7492-7ee7-a72e-15c4cc9ec057")
	if err != nil {
		return err
	}

	mocks := []model.Payload{
		&model.FolderShared{
			FolderID:   "12312331",
			FolderName: "mrk.jpg",
			ActorID:    "32190321093281",
			ActorName:  "marko",
		},
		&model.FileShared{
			FileID:    "55512331",
			FileName:  "report.pdf",
			ActorID:   "32190321093281",
			ActorName: "marko",
		},
		&model.FileAddedToSharedFolder{
			FileID:           "77712331",
			FileName:         "notes.txt",
			SharedFolderID:   "12312331",
			SharedFolderName: "mrk.jpg",
			ActorID:          "32190321093281",
			ActorName:        "marko",
		},
		&model.SharedFileNewVersion{
			FileID:    "55512331",
			FileName:  "report.pdf",
			ActorID:   "32190321093281",
			ActorName: "marko",
		},
	}

	for _, p := range mocks {
		if _, err := r.InsertNotificationsByUserID(ctx, userID, p); err != nil {
			return err
		}
	}

	return nil
}

func decodePayload(typ model.NotificationEventType, raw string) (model.Payload, error) {
	switch typ {
	case model.TypeFolderShared:
		var p model.FolderShared
		err := json.Unmarshal([]byte(raw), &p)
		return p, err
	case model.TypeFileShared:
		var p model.FileShared
		err := json.Unmarshal([]byte(raw), &p)
		return p, err
	case model.TypeFileAddedToSharedFolder:
		var p model.FileAddedToSharedFolder
		err := json.Unmarshal([]byte(raw), &p)
		return p, err
	case model.TypeSharedFileNewVersion:
		var p model.SharedFileNewVersion
		err := json.Unmarshal([]byte(raw), &p)
		return p, err
	default:
		return nil, fmt.Errorf("error decoding payload of type: %q", typ)
	}
}
