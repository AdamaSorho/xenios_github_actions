package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// MeasurementRepository defines the interface for measurement persistence.
type MeasurementRepository interface {
	// CreateBatch inserts multiple measurements in a single operation.
	CreateBatch(ctx context.Context, measurements []*entities.Measurement) ([]*entities.Measurement, error)

	// FindByArtifactID retrieves all measurements linked to a specific artifact.
	FindByArtifactID(ctx context.Context, artifactID string) ([]*entities.Measurement, error)

	// FindByClientID retrieves measurements for a client, ordered by measured_at desc.
	FindByClientID(ctx context.Context, clientID string) ([]*entities.Measurement, error)
}
