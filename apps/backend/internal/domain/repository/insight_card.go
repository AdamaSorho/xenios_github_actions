package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// InsightCardFilter holds query parameters for listing insight cards.
type InsightCardFilter struct {
	CoachID  string
	ClientID string
	Status   string
	Limit    int
	Offset   int
}

// InsightCardRepository defines the interface for managing insight cards.
type InsightCardRepository interface {
	// FindByID retrieves a single insight card by its ID.
	FindByID(ctx context.Context, id string) (*entities.InsightCard, error)

	// ListByCoach retrieves insight cards for a coach with optional filters.
	ListByCoach(ctx context.Context, filter InsightCardFilter) ([]*entities.InsightCard, int, error)

	// ListByClient retrieves insight cards for a specific client with optional status filter.
	ListByClient(ctx context.Context, filter InsightCardFilter) ([]*entities.InsightCard, int, error)

	// Update persists changes to an existing insight card.
	Update(ctx context.Context, insight *entities.InsightCard) error

	// Create stores a new insight card.
	Create(ctx context.Context, insight *entities.InsightCard) error
}
