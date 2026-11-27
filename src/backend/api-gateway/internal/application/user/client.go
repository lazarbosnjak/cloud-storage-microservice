package user

import "context"

type Storage struct {
	Quota     int64
	Used      int64
	Available int64
}

type Client interface {
	Ping(ctx context.Context) (string, error)
	GetStorage(ctx context.Context, userID string) (*Storage, error)
}
