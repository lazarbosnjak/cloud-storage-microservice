package storage

import (
	"context"
	"errors"
	"testing"

	"user-service/internal/domain"
)

type mockRepository struct {
	info        domain.StorageInfo
	err         error
	userID      string
	delta       int64
	getCalls    int
	adjustCalls int
}

func (r *mockRepository) GetStorage(
	ctx context.Context,
	userID string,
) (domain.StorageInfo, error) {
	r.userID = userID
	r.getCalls++

	return r.info, r.err
}

func (r *mockRepository) AdjustStorageUsage(
	ctx context.Context,
	userID string,
	delta int64,
) (domain.StorageInfo, error) {
	r.userID = userID
	r.delta = delta
	r.adjustCalls++

	return r.info, r.err
}

func TestGetStorage(t *testing.T) {
	repository := &mockRepository{
		info: domain.StorageInfo{
			Quota: 1000,
			Used:  200,
		},
	}

	handler := NewHandler(repository)

	result, err := handler.GetStorage(
		context.Background(),
		"user-123",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != repository.info {
		t.Errorf("unexpected storage info: %+v", result)
	}

	if repository.userID != "user-123" {
		t.Errorf("unexpected user ID: %s", repository.userID)
	}
}

func TestGetStorageInvalidUserID(t *testing.T) {
	repository := &mockRepository{}
	handler := NewHandler(repository)

	_, err := handler.GetStorage(context.Background(), " ")

	if !errors.Is(err, ErrInvalidUserID) {
		t.Errorf("expected ErrInvalidUserID, got %v", err)
	}

	if repository.getCalls != 0 {
		t.Error("repository should not be called")
	}
}

func TestAdjustStorageUsage(t *testing.T) {
	for _, delta := range []int64{200, -100} {
		repository := &mockRepository{
			info: domain.StorageInfo{
				Quota: 1000,
				Used:  400 + delta,
			},
		}

		handler := NewHandler(repository)

		result, err := handler.AdjustStorageUsage(
			context.Background(),
			"user-123",
			delta,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if repository.delta != delta {
			t.Errorf("expected delta %d, got %d", delta, repository.delta)
		}

		if result.Used != 400+delta {
			t.Errorf("unexpected storage usage: %d", result.Used)
		}

		if repository.adjustCalls != 1 {
			t.Error("expected one adjustment")
		}
	}
}

func TestAdjustStorageUsageZeroDelta(t *testing.T) {
	repository := &mockRepository{}
	handler := NewHandler(repository)

	_, err := handler.AdjustStorageUsage(
		context.Background(),
		"user-123",
		0,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repository.getCalls != 1 || repository.adjustCalls != 0 {
		t.Error("zero delta should only read storage")
	}
}

func TestAdjustStorageUsageInvalidUserID(t *testing.T) {
	repository := &mockRepository{}
	handler := NewHandler(repository)

	_, err := handler.AdjustStorageUsage(
		context.Background(),
		"",
		100,
	)

	if !errors.Is(err, ErrInvalidUserID) {
		t.Errorf("expected ErrInvalidUserID, got %v", err)
	}

	if repository.adjustCalls != 0 {
		t.Error("repository should not be called")
	}
}

func TestAdjustStorageUsageRepositoryErrors(t *testing.T) {
	testErrors := []error{
		domain.ErrUserNotFound,
		domain.ErrInsufficientStorage,
		domain.ErrInvalidStorageUsage,
		domain.ErrStorageNotInitialized,
	}

	for _, expectedErr := range testErrors {
		t.Run(expectedErr.Error(), func(t *testing.T) {
			repository := &mockRepository{
				err: expectedErr,
			}

			handler := NewHandler(repository)

			_, err := handler.AdjustStorageUsage(
				context.Background(),
				"user-123",
				100,
			)

			if !errors.Is(err, expectedErr) {
				t.Errorf("expected %v, got %v", expectedErr, err)
			}
		})
	}
}
