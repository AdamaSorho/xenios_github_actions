package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// UserRepository defines the interface for user persistence.
type UserRepository interface {
	Create(ctx context.Context, email, passwordHash, name, role string) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByID(ctx context.Context, id string) (*entities.User, error)
}
