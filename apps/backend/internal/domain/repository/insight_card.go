package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// InsightCardRepository defines the interface for insight card persistence.
type InsightCardRepository interface {
	// Create persists a new insight card and returns it with generated ID and timestamps.
	Create(ctx context.Context, card *entities.InsightCard) (*entities.InsightCard, error)

	// FindByClientID returns insight cards for a given client with pagination.
	FindByClientID(ctx context.Context, clientID string, limit, offset int) ([]*entities.InsightCard, error)

	// FindByStatus returns insight cards matching a given status with pagination.
	FindByStatus(ctx context.Context, status entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, error)

	// UpdateStatus changes the status of an insight card by ID.
	UpdateStatus(ctx context.Context, id string, status entities.InsightStatus) (*entities.InsightCard, error)

	// ExistsByMeasurementID returns true if an insight card already exists
	// that references the given measurement ID in its evidence.
	ExistsByMeasurementID(ctx context.Context, measurementID string) (bool, error)
}
