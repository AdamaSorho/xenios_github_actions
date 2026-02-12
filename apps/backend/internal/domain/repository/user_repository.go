package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// UserRepository defines the interface for user persistence operations.
// Implementations handle the actual storage mechanism (e.g., PostgreSQL).
type UserRepository interface {
	// Create persists a new user and returns the created user with generated fields (ID, timestamps).
	// Returns an error if the user could not be created (e.g., duplicate email).
	Create(ctx context.Context, user *entities.User) (*entities.User, error)

	// FindByEmail retrieves a user by their email address.
	// Returns nil, nil when no user is found (not an error — needed by registration to check duplicates).
	// Returns nil, error when an infrastructure error occurs.
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
}
