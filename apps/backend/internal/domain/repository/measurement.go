package repository

import (
	"context"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// MeasurementRepository defines the interface for measurement data access.
type MeasurementRepository interface {
	// FindRecentByClientID returns measurements for a client within a time window.
	FindRecentByClientID(ctx context.Context, clientID string, since time.Time) ([]*entities.Measurement, error)

	// FindByClientIDAndType returns measurements for a client filtered by type.
	FindByClientIDAndType(ctx context.Context, clientID string, measurementType string, since time.Time) ([]*entities.Measurement, error)
}
