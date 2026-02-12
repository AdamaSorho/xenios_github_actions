package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// mockUserRepository is a test implementation of the UserRepository interface.
type mockUserRepository struct {
	createUser *entities.User
	createErr  error
	findUser   *entities.User
	findErr    error
}

func (m *mockUserRepository) Create(ctx context.Context, user *entities.User) (*entities.User, error) {
	return m.createUser, m.createErr
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	return m.findUser, m.findErr
}

func TestUserRepository_InterfaceCompliance(t *testing.T) {
	// Assert that mockUserRepository implements UserRepository
	var _ UserRepository = (*mockUserRepository)(nil)
}

func TestUserRepository_Create_Success(t *testing.T) {
	// Arrange
	now := time.Now().UTC()
	expectedUser := &entities.User{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "$2a$10$hash",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	mock := &mockUserRepository{
		createUser: expectedUser,
		createErr:  nil,
	}

	// Act
	user, err := mock.Create(context.Background(), expectedUser)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if user == nil {
		t.Fatal("expected non-nil user")
	}

	if user.ID != expectedUser.ID {
		t.Errorf("expected ID '%s', got '%s'", expectedUser.ID, user.ID)
	}

	if user.Email != expectedUser.Email {
		t.Errorf("expected email '%s', got '%s'", expectedUser.Email, user.Email)
	}

	if user.Name != expectedUser.Name {
		t.Errorf("expected name '%s', got '%s'", expectedUser.Name, user.Name)
	}
}

func TestUserRepository_Create_Error(t *testing.T) {
	// Arrange
	expectedErr := errors.New("duplicate email")
	mock := &mockUserRepository{
		createUser: nil,
		createErr:  expectedErr,
	}

	inputUser := &entities.User{
		Email:        "existing@example.com",
		Name:         "Test",
		PasswordHash: "$2a$10$hash",
	}

	// Act
	user, err := mock.Create(context.Background(), inputUser)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != expectedErr {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}

	if user != nil {
		t.Error("expected nil user when error returned")
	}
}

func TestUserRepository_FindByEmail_Found(t *testing.T) {
	// Arrange
	now := time.Now().UTC()
	expectedUser := &entities.User{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Email:        "found@example.com",
		Name:         "Found User",
		PasswordHash: "$2a$10$hash",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	mock := &mockUserRepository{
		findUser: expectedUser,
		findErr:  nil,
	}

	// Act
	user, err := mock.FindByEmail(context.Background(), "found@example.com")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if user == nil {
		t.Fatal("expected non-nil user")
	}

	if user.Email != "found@example.com" {
		t.Errorf("expected email 'found@example.com', got '%s'", user.Email)
	}
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	// Arrange - returns nil, nil when user not found (not an error)
	mock := &mockUserRepository{
		findUser: nil,
		findErr:  nil,
	}

	// Act
	user, err := mock.FindByEmail(context.Background(), "notfound@example.com")

	// Assert
	if err != nil {
		t.Fatalf("expected no error for not-found user, got: %v", err)
	}

	if user != nil {
		t.Error("expected nil user for not-found email")
	}
}

func TestUserRepository_FindByEmail_Error(t *testing.T) {
	// Arrange
	expectedErr := errors.New("connection refused")
	mock := &mockUserRepository{
		findUser: nil,
		findErr:  expectedErr,
	}

	// Act
	user, err := mock.FindByEmail(context.Background(), "test@example.com")

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != expectedErr {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}

	if user != nil {
		t.Error("expected nil user when error returned")
	}
}

func TestUserRepository_Create_ContextCancellation(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	now := time.Now().UTC()
	mock := &mockUserRepository{
		createUser: &entities.User{
			ID:        "test-id",
			Email:     "test@example.com",
			Name:      "Test",
			CreatedAt: now,
			UpdatedAt: now,
		},
		createErr: nil,
	}

	inputUser := &entities.User{
		Email:        "test@example.com",
		Name:         "Test",
		PasswordHash: "$2a$10$hash",
	}

	// Act - the mock doesn't check context, but this validates the interface accepts it
	user, err := mock.Create(ctx, inputUser)

	// Assert
	if err != nil {
		t.Fatalf("mock doesn't check context, expected no error, got: %v", err)
	}

	if user == nil {
		t.Fatal("expected non-nil user")
	}
}
