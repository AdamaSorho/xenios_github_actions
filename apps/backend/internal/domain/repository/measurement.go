package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// MeasurementRepository defines the interface for measurement persistence.
type MeasurementRepository interface {
	// CreateBatch stores multiple measurements in a single operation.
	CreateBatch(ctx context.Context, measurements []*entities.Measurement) error
}
