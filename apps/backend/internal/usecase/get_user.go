package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetUserUseCase handles the business logic for retrieving a user.
type GetUserUseCase struct {
	userRepo repository.UserRepository
}

// NewGetUserUseCase creates a new GetUserUseCase with dependency injection.
func NewGetUserUseCase(userRepo repository.UserRepository) *GetUserUseCase {
	return &GetUserUseCase{userRepo: userRepo}
}

// Execute retrieves a user by ID.
func (uc *GetUserUseCase) Execute(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	return uc.userRepo.FindByID(ctx, id)
}
