package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

func TestCreateUserUseCase_Execute_Success(t *testing.T) {
	mockRepo := &MockUserRepository{
		findByEmailFunc: func(ctx context.Context, email string) (*entities.User, error) {
			return nil, nil // Email not found
		},
		createFunc: func(ctx context.Context, user *entities.User) error {
			return nil
		},
	}

	uc := usecase.NewCreateUserUseCase(mockRepo)
	user, err := uc.Execute(context.Background(), usecase.CreateUserInput{
		Email: "test@example.com",
		Name:  "Test User",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %v", user.Email)
	}
	if user.Name != "Test User" {
		t.Errorf("expected name Test User, got %v", user.Name)
	}
}

func TestCreateUserUseCase_Execute_EmailAlreadyExists(t *testing.T) {
	mockRepo := &MockUserRepository{
		findByEmailFunc: func(ctx context.Context, email string) (*entities.User, error) {
			return &entities.User{Email: email}, nil // Email exists
		},
	}

	uc := usecase.NewCreateUserUseCase(mockRepo)
	user, err := uc.Execute(context.Background(), usecase.CreateUserInput{
		Email: "existing@example.com",
		Name:  "Test User",
	})

	if err != usecase.ErrEmailAlreadyExists {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
}

func TestCreateUserUseCase_Execute_InvalidEmail(t *testing.T) {
	mockRepo := &MockUserRepository{}

	uc := usecase.NewCreateUserUseCase(mockRepo)
	user, err := uc.Execute(context.Background(), usecase.CreateUserInput{
		Email: "",
		Name:  "Test User",
	})

	if err != usecase.ErrInvalidEmail {
		t.Errorf("expected ErrInvalidEmail, got %v", err)
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
}

func TestCreateUserUseCase_Execute_InvalidName(t *testing.T) {
	mockRepo := &MockUserRepository{}

	uc := usecase.NewCreateUserUseCase(mockRepo)
	user, err := uc.Execute(context.Background(), usecase.CreateUserInput{
		Email: "test@example.com",
		Name:  "",
	})

	if err != usecase.ErrInvalidName {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
}

func TestCreateUserUseCase_Execute_RepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")
	mockRepo := &MockUserRepository{
		findByEmailFunc: func(ctx context.Context, email string) (*entities.User, error) {
			return nil, nil
		},
		createFunc: func(ctx context.Context, user *entities.User) error {
			return expectedErr
		},
	}

	uc := usecase.NewCreateUserUseCase(mockRepo)
	user, err := uc.Execute(context.Background(), usecase.CreateUserInput{
		Email: "test@example.com",
		Name:  "Test User",
	})

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}
}
