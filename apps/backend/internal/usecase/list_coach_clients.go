package usecase

import (
	"context"
	"errors"

	"github.com/xenios/backend/internal/domain"
)

// ListCoachClientsUseCase handles listing clients for a coach.
type ListCoachClientsUseCase struct {
	repo domain.CoachClientRepository
}

// NewListCoachClientsUseCase creates a new ListCoachClientsUseCase instance.
func NewListCoachClientsUseCase(repo domain.CoachClientRepository) *ListCoachClientsUseCase {
	return &ListCoachClientsUseCase{repo: repo}
}

// ListCoachClientsInput holds the input data for listing coach clients.
type ListCoachClientsInput struct {
	CoachID string
	Limit   int
	Offset  int
}

// Validate checks that the input is valid.
func (i *ListCoachClientsInput) Validate() error {
	if i.CoachID == "" {
		return errors.New("coach_id is required")
	}
	if i.Limit < 0 {
		return errors.New("limit must be non-negative")
	}
	if i.Offset < 0 {
		return errors.New("offset must be non-negative")
	}
	return nil
}

const defaultLimit = 20
const maxLimit = 100

// Execute returns the list of clients for the given coach.
func (uc *ListCoachClientsUseCase) Execute(ctx context.Context, input ListCoachClientsInput) ([]*domain.CoachClient, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	limit := input.Limit
	if limit == 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	return uc.repo.ListByCoachID(ctx, input.CoachID, limit, input.Offset)
}
