package repository

import (
	"context"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// MeasurementRepository defines the interface for persisting nutrition measurements.
type MeasurementRepository interface {
	// BatchCreate stores multiple measurements in a single operation.
	BatchCreate(ctx context.Context, measurements []*entities.Measurement) error

	// FindByClientAndDateRange returns measurements for a client within a date range.
	FindByClientAndDateRange(ctx context.Context, clientID string, from, to time.Time) ([]*entities.Measurement, error)
}
