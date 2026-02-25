package repository

import (
	"context"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// MeasurementRepository defines persistence operations for wearable measurements.
type MeasurementRepository interface {
	// UpsertBatch inserts measurements, skipping duplicates for the same
	// client_id + source + measurement_type + measured_at combination.
	// Returns the number of rows actually inserted.
	UpsertBatch(ctx context.Context, measurements []entities.Measurement) (int, error)

	// Average computes the mean value for a given metric type within a time window.
	// Returns nil when no data exists for the window.
	Average(ctx context.Context, clientID string, source entities.WearableSource, mt entities.MeasurementType, since time.Time) (*float64, error)
}
