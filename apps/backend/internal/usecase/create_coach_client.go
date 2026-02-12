package usecase

import (
	"context"
	"errors"

	"github.com/xenios/backend/internal/domain"
)

// CreateCoachClientUseCase handles creating a new coach-client relationship.
type CreateCoachClientUseCase struct {
	repo domain.CoachClientRepository
}

// NewCreateCoachClientUseCase creates a new CreateCoachClientUseCase instance.
func NewCreateCoachClientUseCase(repo domain.CoachClientRepository) *CreateCoachClientUseCase {
	return &CreateCoachClientUseCase{repo: repo}
}

// CreateCoachClientInput holds the input data for creating a coach-client relationship.
type CreateCoachClientInput struct {
	CoachID  string
	ClientID string
}

// Validate checks that the input is valid.
func (i *CreateCoachClientInput) Validate() error {
	if i.CoachID == "" {
		return errors.New("coach_id is required")
	}
	if i.ClientID == "" {
		return errors.New("client_id is required")
	}
	if i.CoachID == i.ClientID {
		return errors.New("coach_id and client_id must be different")
	}
	return nil
}

// Execute creates a new coach-client relationship.
func (uc *CreateCoachClientUseCase) Execute(ctx context.Context, input CreateCoachClientInput) (*domain.CoachClient, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	cc := &domain.CoachClient{
		CoachID:  input.CoachID,
		ClientID: input.ClientID,
		Status:   domain.CoachClientStatusActive,
	}

	return uc.repo.Create(ctx, cc)
}
