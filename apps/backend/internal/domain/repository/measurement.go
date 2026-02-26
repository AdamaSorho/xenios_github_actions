package repository

import (
	"context"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// MeasurementRepository defines the interface for measurement data access.
type MeasurementRepository interface {
	FindByClientID(ctx context.Context, clientID string, limit, offset int) ([]*entities.Measurement, error)
	FindByClientIDAndType(ctx context.Context, clientID string, measurementType string, since time.Time) ([]*entities.Measurement, error)
	FindRecentByArtifactID(ctx context.Context, artifactID string) ([]*entities.Measurement, error)
}
