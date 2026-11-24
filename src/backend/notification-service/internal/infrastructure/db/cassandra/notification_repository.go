package cassandra

import (
	"fmt"
	"log/slog"

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

// TODO:
// - Create mock data
// - Fetch mock data
// - Create 2 more tables
// - Create RPC
