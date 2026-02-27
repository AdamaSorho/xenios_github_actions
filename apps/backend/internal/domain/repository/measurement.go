package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// MeasurementRepository defines the interface for managing measurements.
type MeasurementRepository interface {
	// Create stores a new measurement.
	Create(ctx context.Context, m *entities.Measurement) (*entities.Measurement, error)

	// FindByClientID retrieves measurements for a client with filtering and pagination.
	FindByClientID(ctx context.Context, filter entities.MeasurementFilter) (*entities.MeasurementPage, error)

	// FindLatestByClientID returns the most recent measurement for each type for a client.
	FindLatestByClientID(ctx context.Context, clientID string) ([]*entities.Measurement, error)

	// FindByType retrieves measurements of a specific type for a client.
	FindByType(ctx context.Context, clientID, measurementType string) ([]*entities.Measurement, error)
}
