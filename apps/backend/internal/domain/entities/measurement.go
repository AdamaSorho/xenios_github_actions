package entities

import "time"

// MeasurementType represents the type of body composition measurement.
type MeasurementType string

const (
	MeasurementTypeWeight              MeasurementType = "weight"
	MeasurementTypeBodyFatPct          MeasurementType = "body_fat_pct"
	MeasurementTypeSkeletalMuscleMass  MeasurementType = "skeletal_muscle_mass"
	MeasurementTypeBMR                 MeasurementType = "bmr"
	MeasurementTypeTotalBodyWater      MeasurementType = "total_body_water"
	MeasurementTypeLeanBodyMass        MeasurementType = "lean_body_mass"
)

// ValidMeasurementTypes lists all recognized measurement types.
var ValidMeasurementTypes = map[MeasurementType]bool{
	MeasurementTypeWeight:             true,
	MeasurementTypeBodyFatPct:         true,
	MeasurementTypeSkeletalMuscleMass: true,
	MeasurementTypeBMR:                true,
	MeasurementTypeTotalBodyWater:     true,
	MeasurementTypeLeanBodyMass:       true,
}

// IsValidMeasurementType checks if the given type is a known measurement type.
func IsValidMeasurementType(mt MeasurementType) bool {
	return ValidMeasurementTypes[mt]
}

// Measurement represents a single body composition measurement.
type Measurement struct {
	ID              string          `json:"id"`
	ClientID        string          `json:"client_id"`
	RecordedBy      string          `json:"recorded_by"`
	MeasurementType MeasurementType `json:"measurement_type"`
	Value           float64         `json:"value"`
	Unit            string          `json:"unit"`
	MeasuredAt      time.Time       `json:"measured_at"`
	Notes           string          `json:"notes,omitempty"`
	ArtifactID      string          `json:"artifact_id,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

// ExtractionResult holds the outcome of an InBody PDF extraction.
type ExtractionResult struct {
	Measurements []*Measurement `json:"measurements"`
	ArtifactID   string         `json:"artifact_id"`
	Partial      bool           `json:"partial"`
	Errors       []string       `json:"errors,omitempty"`
}
