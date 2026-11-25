package create

import (
	"context"
	"errors"
	"testing"

	"user-service/internal/domain"
)

type mockUserRepository struct {
	user domain.User
	err  error
}

func (r *mockUserRepository) Create(ctx context.Context, user domain.User) error {
	r.user = user
	return r.err
}

func TestRegisterInitializesStorage(t *testing.T) {
	const defaultQuota int64 = 1073741824

	repository := &mockUserRepository{}
	handler := NewHandler(repository, defaultQuota)

	query := Query{
		ID:        "user-123",
		Email:     "test@example.com",
		FirstName: "Petar",
		LastName:  "Petrovic",
	}

	err := handler.Register(context.Background(), query)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := domain.User{
		ID:           query.ID,
		Email:        query.Email,
		FirstName:    query.FirstName,
		LastName:     query.LastName,
		StorageQuota: defaultQuota,
		StorageUsed:  0,
	}

	if repository.user != expected {
		t.Errorf("expected %+v, got %+v", expected, repository.user)
	}
}

func TestRegisterRepositoryError(t *testing.T) {
	expectedErr := errors.New("repository error")

	repository := &mockUserRepository{
		err: expectedErr,
	}

	handler := NewHandler(repository, 1073741824)

	err := handler.Register(context.Background(), Query{
		ID:        "user-123",
		Email:     "test@example.com",
		FirstName: "Petar",
		LastName:  "Petrovic",
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}
