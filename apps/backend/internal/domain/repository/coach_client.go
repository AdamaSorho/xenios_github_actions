package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// CoachClientRepository defines the interface for managing coach-client relationships.
type CoachClientRepository interface {
	// Create stores a new coach-client relationship.
	Create(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error)

	// ListByCoachID retrieves all clients for a given coach with pagination.
	ListByCoachID(ctx context.Context, coachID string, limit, offset int) ([]*entities.CoachClient, error)

	// FindByCoachAndClient checks if a coach-client relationship exists.
	// Returns the relationship if found, nil if not found.
	FindByCoachAndClient(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error)
}
