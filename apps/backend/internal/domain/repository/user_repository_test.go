package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

// mockUserRepository implements UserRepository for testing that the interface
// is correctly defined and can be implemented.
type mockUserRepository struct {
	createFn      func(ctx context.Context, user *entities.User) (*entities.User, error)
	findByEmailFn func(ctx context.Context, email string) (*entities.User, error)
	findByIDFn    func(ctx context.Context, id string) (*entities.User, error)
}

func (m *mockUserRepository) Create(ctx context.Context, user *entities.User) (*entities.User, error) {
	return m.createFn(ctx, user)
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	return m.findByEmailFn(ctx, email)
}

func (m *mockUserRepository) FindByID(ctx context.Context, id string) (*entities.User, error) {
	return m.findByIDFn(ctx, id)
}

func TestUserRepository_InterfaceCompiles(t *testing.T) {
	// This test verifies the UserRepository interface can be implemented.
	var repo UserRepository = &mockUserRepository{
		createFn: func(_ context.Context, user *entities.User) (*entities.User, error) {
			return user, nil
		},
		findByEmailFn: func(_ context.Context, _ string) (*entities.User, error) {
			return nil, nil
		},
		findByIDFn: func(_ context.Context, _ string) (*entities.User, error) {
			return nil, nil
		},
	}

	if repo == nil {
		t.Fatal("expected non-nil repository")
	}
}

func TestUserRepository_Create_ReturnsUser(t *testing.T) {
	user, err := entities.NewUser("test@example.com", "Test User", "$2a$10$hash")
	if err != nil {
		t.Fatalf("failed to create user entity: %v", err)
	}

	repo := &mockUserRepository{
		createFn: func(_ context.Context, u *entities.User) (*entities.User, error) {
			return u, nil
		},
	}

	result, err := repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Email != user.Email {
		t.Errorf("expected email '%s', got '%s'", user.Email, result.Email)
	}
}

func TestUserRepository_FindByEmail_ReturnsNilForNotFound(t *testing.T) {
	repo := &mockUserRepository{
		findByEmailFn: func(_ context.Context, _ string) (*entities.User, error) {
			return nil, nil
		},
	}

	result, err := repo.FindByEmail(context.Background(), "nonexistent@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != nil {
		t.Error("expected nil result for non-existent user")
	}
}

func TestUserRepository_FindByID_ReturnsNilForNotFound(t *testing.T) {
	repo := &mockUserRepository{
		findByIDFn: func(_ context.Context, _ string) (*entities.User, error) {
			return nil, nil
		},
	}

	result, err := repo.FindByID(context.Background(), "nonexistent-id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != nil {
		t.Error("expected nil result for non-existent user")
	}
}
