package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// NutritionRepository defines the interface for nutrition data persistence.
type NutritionRepository interface {
	// BatchCreateMeasurements stores multiple nutrition measurements at once.
	BatchCreateMeasurements(ctx context.Context, measurements []*entities.Measurement) error

	// StoreAverages stores computed nutrition averages for a client.
	StoreAverages(ctx context.Context, averages []*entities.NutritionAverage) error

	// FindMeasurementsByClientAndType retrieves measurements for a client filtered by type.
	FindMeasurementsByClientAndType(ctx context.Context, clientID, measurementType string, limit int) ([]*entities.Measurement, error)

	// FindAveragesByClient retrieves nutrition averages for a client.
	FindAveragesByClient(ctx context.Context, clientID string) ([]*entities.NutritionAverage, error)
}
