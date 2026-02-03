package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// MockUserRepository is a test double for UserRepository.
type MockUserRepository struct {
	findByIDFunc    func(ctx context.Context, id uuid.UUID) (*entities.User, error)
	findByEmailFunc func(ctx context.Context, email string) (*entities.User, error)
	createFunc      func(ctx context.Context, user *entities.User) error
	updateFunc      func(ctx context.Context, user *entities.User) error
	deleteFunc      func(ctx context.Context, id uuid.UUID) error
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	if m.findByEmailFunc != nil {
		return m.findByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func TestGetUserUseCase_Execute_Success(t *testing.T) {
	userID := uuid.New()
	expectedUser := &entities.User{
		ID:        userID,
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo := &MockUserRepository{
		findByIDFunc: func(ctx context.Context, id uuid.UUID) (*entities.User, error) {
			if id == userID {
				return expectedUser, nil
			}
			return nil, nil
		},
	}

	uc := usecase.NewGetUserUseCase(mockRepo)
	user, err := uc.Execute(context.Background(), userID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.ID != userID {
		t.Errorf("expected ID %v, got %v", userID, user.ID)
	}
}

func TestGetUserUseCase_Execute_NotFound(t *testing.T) {
	mockRepo := &MockUserRepository{
		findByIDFunc: func(ctx context.Context, id uuid.UUID) (*entities.User, error) {
			return nil, nil
		},
	}

	uc := usecase.NewGetUserUseCase(mockRepo)
	user, err := uc.Execute(context.Background(), uuid.New())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
}

func TestGetUserUseCase_Execute_Error(t *testing.T) {
	expectedErr := errors.New("database error")
	mockRepo := &MockUserRepository{
		findByIDFunc: func(ctx context.Context, id uuid.UUID) (*entities.User, error) {
			return nil, expectedErr
		},
	}

	uc := usecase.NewGetUserUseCase(mockRepo)
	user, err := uc.Execute(context.Background(), uuid.New())

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
}
