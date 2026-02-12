package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// UserRepository defines the interface for user persistence operations.
// FindByEmail returns nil, nil when no user is found (not an error).
type UserRepository interface {
	Create(ctx context.Context, user *entities.User) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
}
