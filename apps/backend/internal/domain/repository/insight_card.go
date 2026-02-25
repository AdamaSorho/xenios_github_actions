package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// InsightCardRepository defines the interface for insight card persistence.
type InsightCardRepository interface {
	// Create stores a new insight card and returns the created record.
	Create(ctx context.Context, card *entities.InsightCard) (*entities.InsightCard, error)

	// FindByClientID returns all insight cards for a given client.
	FindByClientID(ctx context.Context, clientID string) ([]*entities.InsightCard, error)

	// FindByStatus returns all insight cards with the given status.
	FindByStatus(ctx context.Context, status entities.InsightStatus) ([]*entities.InsightCard, error)

	// UpdateStatus updates the status of an insight card.
	UpdateStatus(ctx context.Context, id string, status entities.InsightStatus) (*entities.InsightCard, error)

	// ExistsByEvidence checks if an insight card already exists that references
	// the given measurement ID, preventing duplicate insight generation.
	ExistsByEvidence(ctx context.Context, clientID string, measurementID string) (bool, error)
}
