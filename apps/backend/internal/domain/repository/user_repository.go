package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// UserRepository defines the interface for user persistence operations.
// Concrete implementations reside in outer layers (e.g., PostgresUserRepository).
type UserRepository interface {
	// Create persists a new user and returns the created user with a generated ID.
	// Returns ErrDuplicateEmail if a user with the same email already exists.
	Create(ctx context.Context, user *entities.User) (*entities.User, error)

	// FindByEmail retrieves a user by email address.
	// Returns nil and no error if the user is not found.
	FindByEmail(ctx context.Context, email string) (*entities.User, error)

	// FindByID retrieves a user by their unique identifier.
	// Returns nil and no error if the user is not found.
	FindByID(ctx context.Context, id string) (*entities.User, error)
}
