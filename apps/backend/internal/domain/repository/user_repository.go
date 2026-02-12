package repository

import (
	"context"
	"errors"

	"github.com/xenios/backend/internal/domain/entities"
)

// ErrDuplicateEmail is returned when attempting to create a user with an email
// that already exists in the database.
var ErrDuplicateEmail = errors.New("email already exists")

// UserRepository defines the interface for user persistence operations.
type UserRepository interface {
	// Create inserts a new user and returns the created user with
	// database-generated fields (id, created_at, updated_at).
	Create(ctx context.Context, user *entities.User) (*entities.User, error)

	// FindByEmail retrieves a user by their email address.
	// Returns nil, nil when no user is found with the given email.
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
}
