package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// MeasurementRepository defines the interface for measurement persistence.
type MeasurementRepository interface {
	// CreateBatch stores multiple measurements atomically.
	CreateBatch(ctx context.Context, measurements []*entities.Measurement) ([]*entities.Measurement, error)

	// FindByArtifactID retrieves all measurements linked to a specific artifact.
	FindByArtifactID(ctx context.Context, artifactID string) ([]*entities.Measurement, error)
}
