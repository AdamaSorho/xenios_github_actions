package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// MeasurementRepository defines the interface for measurement persistence.
type MeasurementRepository interface {
	CreateBatch(ctx context.Context, measurements []*entities.Measurement) ([]*entities.Measurement, error)
	FindByArtifactID(ctx context.Context, artifactID string) ([]*entities.Measurement, error)
}
