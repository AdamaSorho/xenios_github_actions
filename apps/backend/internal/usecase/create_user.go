package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ErrDuplicateEmail is returned when attempting to create a user with an email
// that already exists in the system.
var ErrDuplicateEmail = errors.New("a user with this email already exists")

// CreateUserInput contains the data required to create a new user.
type CreateUserInput struct {
	Email        string
	Name         string
	PasswordHash string
}

// CreateUserUseCase handles the creation of new user accounts.
type CreateUserUseCase struct {
	userRepo repository.UserRepository
}

// NewCreateUserUseCase creates a new CreateUserUseCase with the given repository.
func NewCreateUserUseCase(userRepo repository.UserRepository) *CreateUserUseCase {
	return &CreateUserUseCase{userRepo: userRepo}
}

// Execute creates a new user after validating input and checking for duplicates.
// Returns the created user or an error if validation fails, the email is taken,
// or the repository encounters an error.
func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*entities.User, error) {
	// Create and validate the user entity
	user, err := entities.NewUser(input.Email, input.Name, input.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("invalid user data: %w", err)
	}

	// Check for duplicate email
	existing, err := uc.userRepo.FindByEmail(ctx, user.Email)
	if err != nil {
		return nil, fmt.Errorf("checking for existing user: %w", err)
	}
	if existing != nil {
		return nil, ErrDuplicateEmail
	}

	// Persist the user
	created, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	return created, nil
}
