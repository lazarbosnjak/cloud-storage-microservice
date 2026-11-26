package grpc

import (
	"context"
	"sync"

	"file-service/internal/domain/quota"
)

type fakeQuotaClient struct {
	mu    sync.Mutex
	limit int64
	used  map[string]int64
}

var _ quota.Client = (*fakeQuotaClient)(nil)

func newFakeQuotaClient() *fakeQuotaClient {
	return &fakeQuotaClient{
		limit: 1073741824,
		used:  make(map[string]int64),
	}
}

func (c *fakeQuotaClient) GetStorage(
	ctx context.Context,
	userID string,
) (quota.StorageInfo, error) {
	if err := ctx.Err(); err != nil {
		return quota.StorageInfo{}, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return quota.StorageInfo{
		Quota:     c.limit,
		Used:      c.used[userID],
		Available: c.limit - c.used[userID],
	}, nil
}

func (c *fakeQuotaClient) AdjustStorageUsage(
	ctx context.Context,
	userID string,
	delta int64,
) (quota.StorageInfo, error) {
	if err := ctx.Err(); err != nil {
		return quota.StorageInfo{}, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	current := c.used[userID]

	if delta > c.limit-current {
		return quota.StorageInfo{}, quota.ErrInsufficientStorage
	}

	if delta < -current {
		return quota.StorageInfo{}, quota.ErrInvalidStorageUsage
	}

	used := current + delta
	c.used[userID] = used

	return quota.StorageInfo{
		Quota:     c.limit,
		Used:      used,
		Available: c.limit - used,
	}, nil
}
