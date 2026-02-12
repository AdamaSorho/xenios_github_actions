package usecase

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// ListCoachClientsUseCase handles listing clients for a coach.
type ListCoachClientsUseCase struct {
	repo repository.CoachClientRepository
}

// NewListCoachClientsUseCase creates a new ListCoachClientsUseCase.
func NewListCoachClientsUseCase(repo repository.CoachClientRepository) *ListCoachClientsUseCase {
	return &ListCoachClientsUseCase{repo: repo}
}

// Execute lists clients for a coach with pagination.
func (uc *ListCoachClientsUseCase) Execute(ctx context.Context, coachID string, limit, offset int) ([]*entities.CoachClient, error) {
	if coachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if limit < 0 {
		return nil, &ValidationError{Message: "limit must be non-negative"}
	}
	if offset < 0 {
		return nil, &ValidationError{Message: "offset must be non-negative"}
	}

	if limit == 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	return uc.repo.ListByCoachID(ctx, coachID, limit, offset)
}
