package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// InsightCardRepository defines the interface for managing insight cards.
type InsightCardRepository interface {
	// Create stores a new insight card.
	Create(ctx context.Context, card *entities.InsightCard) (*entities.InsightCard, error)

	// FindByID retrieves an insight card by its ID.
	FindByID(ctx context.Context, id string) (*entities.InsightCard, error)

	// Update persists changes to an existing insight card.
	Update(ctx context.Context, card *entities.InsightCard) (*entities.InsightCard, error)

	// ListByCoachID retrieves insight cards for a coach, filtered by optional status.
	ListByCoachID(ctx context.Context, filter entities.InsightQueryFilter) ([]*entities.InsightCard, int, error)

	// ListByClientID retrieves insight cards for a client, filtered by optional status.
	ListByClientID(ctx context.Context, filter entities.InsightQueryFilter) ([]*entities.InsightCard, int, error)
}
