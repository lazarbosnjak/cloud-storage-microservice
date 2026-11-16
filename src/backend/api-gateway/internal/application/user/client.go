package user

import "context"

type Client interface {
	Ping(ctx context.Context) (string, error)
}
