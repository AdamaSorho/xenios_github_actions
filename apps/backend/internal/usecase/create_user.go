package usecase

import (
	"context"
	"errors"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidName        = errors.New("name is required")
)

// CreateUserUseCase handles the business logic for creating a user.
type CreateUserUseCase struct {
	userRepo repository.UserRepository
}

// NewCreateUserUseCase creates a new CreateUserUseCase with dependency injection.
func NewCreateUserUseCase(userRepo repository.UserRepository) *CreateUserUseCase {
	return &CreateUserUseCase{userRepo: userRepo}
}

// CreateUserInput contains the data needed to create a user.
type CreateUserInput struct {
	Email string
	Name  string
}

// Execute creates a new user after validating business rules.
func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*entities.User, error) {
	if input.Email == "" {
		return nil, ErrInvalidEmail
	}
	if input.Name == "" {
		return nil, ErrInvalidName
	}

	// Check if email already exists
	existing, _ := uc.userRepo.FindByEmail(ctx, input.Email)
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	user := entities.NewUser(input.Email, input.Name)
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
