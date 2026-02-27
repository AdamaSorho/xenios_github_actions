package repository

import (
	"context"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// MeasurementRepository defines operations for persisting health measurements.
type MeasurementRepository interface {
	// UpsertBatch inserts measurements, skipping duplicates (same client_id, measurement_type, measured_at, source).
	UpsertBatch(ctx context.Context, measurements []entities.Measurement) (inserted int, err error)
	// Average computes the average value of a measurement type for a client within a time window.
	Average(ctx context.Context, clientID string, measurementType entities.MeasurementType, source entities.WearableSource, since time.Time) (float64, error)
}
