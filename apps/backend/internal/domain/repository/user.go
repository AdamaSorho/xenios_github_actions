package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/xenios/backend/internal/domain/entities"
)

// UserRepository defines the interface for user data access.
// NOTE: This is an INTERFACE only - no database imports here!
// Implementations live in the adapter/repository layer.
type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	Create(ctx context.Context, user *entities.User) error
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}
