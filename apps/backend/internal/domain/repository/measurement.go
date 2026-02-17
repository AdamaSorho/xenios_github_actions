package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// MeasurementRepository defines the interface for measurement persistence.
type MeasurementRepository interface {
	// Create stores a new measurement and returns the created measurement with a generated ID.
	Create(ctx context.Context, m *entities.Measurement) (*entities.Measurement, error)

	// BatchCreate stores multiple measurements in a single operation.
	BatchCreate(ctx context.Context, measurements []*entities.Measurement) ([]*entities.Measurement, error)

	// FindByArtifactID returns all measurements linked to a given artifact.
	FindByArtifactID(ctx context.Context, artifactID string) ([]*entities.Measurement, error)
}
