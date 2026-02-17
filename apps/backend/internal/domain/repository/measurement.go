package repository

import (
	"context"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// MeasurementRepository defines the interface for wearable measurement persistence.
type MeasurementRepository interface {
	// BulkUpsert inserts measurements, skipping duplicates for the same client+source+type+date.
	BulkUpsert(ctx context.Context, measurements []entities.WearableMeasurement) (inserted int, err error)

	// GetAverages returns the average value for a measurement type within a time window.
	GetAverages(ctx context.Context, clientID string, source entities.WearableSource, measurementType entities.MeasurementType, from time.Time, to time.Time) (float64, error)
}
