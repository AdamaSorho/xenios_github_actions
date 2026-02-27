package repository

import (
	"context"
	"time"
)

// Measurement represents a stored measurement entry.
type Measurement struct {
	ID              string
	ClientID        string
	RecordedBy      string
	MeasurementType string
	Value           float64
	Unit            string
	MeasuredAt      time.Time
	ArtifactID      string
	Notes           string
	CreatedAt       time.Time
}

// MeasurementRepository defines the interface for measurement persistence.
type MeasurementRepository interface {
	BatchCreate(ctx context.Context, measurements []Measurement) error
}
