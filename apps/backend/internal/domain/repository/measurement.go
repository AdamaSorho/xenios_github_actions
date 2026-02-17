package repository

import (
	"context"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// MeasurementRepository defines the interface for querying measurements.
type MeasurementRepository interface {
	// FindByClientID returns measurements for a client within a time range.
	FindByClientID(ctx context.Context, clientID string, from, to time.Time) ([]*entities.Measurement, error)

	// FindByArtifactID returns measurements extracted from a specific artifact.
	FindByArtifactID(ctx context.Context, artifactID string) ([]*entities.Measurement, error)
}
