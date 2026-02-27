package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// MeasurementInput holds the data needed to create a measurement.
type MeasurementInput struct {
	ClientID        string
	RecordedBy      string
	MeasurementType string
	Value           float64
	Unit            string
	ArtifactID      *string
	ReferenceLow    *float64
	ReferenceHigh   *float64
	Flag            *entities.LabFlag
	Notes           string
}

// MeasurementRepository defines the interface for measurement persistence.
type MeasurementRepository interface {
	// CreateBatch stores multiple measurements atomically.
	CreateBatch(ctx context.Context, measurements []MeasurementInput) (int, error)
}
