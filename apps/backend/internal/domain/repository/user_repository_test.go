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
	createUser    *entities.User
	createErr     error
	findByEmailUser *entities.User
	findByEmailErr  error
}

func (m *mockUserRepository) Create(ctx context.Context, user *entities.User) (*entities.User, error) {
	return m.createUser, m.createErr
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	return m.findByEmailUser, m.findByEmailErr
}

func TestUserRepository_InterfaceCompliance(t *testing.T) {
	// Assert that mockUserRepository implements UserRepository
	var _ UserRepository = (*mockUserRepository)(nil)
}

func TestUserRepository_MockCreate_ReturnsUser(t *testing.T) {
	// Arrange
	now := time.Now()
	mock := &mockUserRepository{
		createUser: &entities.User{
			ID:           "created-id",
			Email:        "new@example.com",
			Name:         "New User",
			PasswordHash: "$2a$10$hash",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		createErr: nil,
	}

	inputUser := &entities.User{
		Email:        "new@example.com",
		Name:         "New User",
		PasswordHash: "$2a$10$hash",
	}

	// Act
	result, err := mock.Create(context.Background(), inputUser)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil user")
	}
	if result.ID != "created-id" {
		t.Errorf("expected ID 'created-id', got '%s'", result.ID)
	}
	if result.Email != "new@example.com" {
		t.Errorf("expected email 'new@example.com', got '%s'", result.Email)
	}
}

func TestUserRepository_MockCreate_ReturnsError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("database error")
	mock := &mockUserRepository{
		createUser: nil,
		createErr:  expectedErr,
	}

	// Act
	result, err := mock.Create(context.Background(), &entities.User{})

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != expectedErr {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
	if result != nil {
		t.Error("expected nil user when error returned")
	}
}

func TestUserRepository_MockFindByEmail_ReturnsUser(t *testing.T) {
	// Arrange
	now := time.Now()
	mock := &mockUserRepository{
		findByEmailUser: &entities.User{
			ID:        "found-id",
			Email:     "exists@example.com",
			Name:      "Existing User",
			CreatedAt: now,
			UpdatedAt: now,
		},
		findByEmailErr: nil,
	}

	// Act
	result, err := mock.FindByEmail(context.Background(), "exists@example.com")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil user")
	}
	if result.Email != "exists@example.com" {
		t.Errorf("expected email 'exists@example.com', got '%s'", result.Email)
	}
}

func TestUserRepository_MockFindByEmail_ReturnsNilWhenNotFound(t *testing.T) {
	// Arrange - nil, nil means user not found (not an error)
	mock := &mockUserRepository{
		findByEmailUser: nil,
		findByEmailErr:  nil,
	}

	// Act
	result, err := mock.FindByEmail(context.Background(), "notfound@example.com")

	// Assert
	if err != nil {
		t.Fatalf("expected no error for not-found, got: %v", err)
	}
	if result != nil {
		t.Error("expected nil user for not-found email")
	}
}

func TestUserRepository_MockFindByEmail_ReturnsError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("connection refused")
	mock := &mockUserRepository{
		findByEmailUser: nil,
		findByEmailErr:  expectedErr,
	}

	// Act
	result, err := mock.FindByEmail(context.Background(), "any@example.com")

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != expectedErr {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
	if result != nil {
		t.Error("expected nil user when error returned")
	}
}
