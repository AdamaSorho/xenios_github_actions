package usecase

import (
	"context"
	"errors"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// CreateCoachClientUseCase handles creating coach-client relationships.
type CreateCoachClientUseCase struct {
	repo repository.CoachClientRepository
}

// NewCreateCoachClientUseCase creates a new CreateCoachClientUseCase.
func NewCreateCoachClientUseCase(repo repository.CoachClientRepository) *CreateCoachClientUseCase {
	return &CreateCoachClientUseCase{repo: repo}
}

// Execute creates a new coach-client relationship after validation.
func (uc *CreateCoachClientUseCase) Execute(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
	if coachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if clientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}
	if coachID == clientID {
		return nil, &ValidationError{Message: "coach_id and client_id must be different"}
	}

	return uc.repo.Create(ctx, coachID, clientID)
}

// ValidationError represents a validation failure in the use case layer.
// These errors are safe to expose to API consumers.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// IsValidationError checks whether the given error is a ValidationError.
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}

// AuthorizationError represents an authorization failure in the use case layer.
type AuthorizationError struct {
	Message string
}

func (e *AuthorizationError) Error() string {
	return e.Message
}

// IsAuthorizationError checks whether the given error is an AuthorizationError.
func IsAuthorizationError(err error) bool {
	var ae *AuthorizationError
	return errors.As(err, &ae)
}
