package entities

import "time"

// MeasurementType represents the type of body measurement.
type MeasurementType string

const (
	MeasurementTypeWeight             MeasurementType = "weight"
	MeasurementTypeBodyFatPct         MeasurementType = "body_fat_pct"
	MeasurementTypeSkeletalMuscleMass MeasurementType = "skeletal_muscle_mass"
	MeasurementTypeBMR                MeasurementType = "bmr"
	MeasurementTypeTotalBodyWater     MeasurementType = "total_body_water"
	MeasurementTypeLeanBodyMass       MeasurementType = "lean_body_mass"
)

// ValidMeasurementTypes enumerates all known InBody measurement types.
var ValidMeasurementTypes = map[MeasurementType]bool{
	MeasurementTypeWeight:             true,
	MeasurementTypeBodyFatPct:         true,
	MeasurementTypeSkeletalMuscleMass: true,
	MeasurementTypeBMR:                true,
	MeasurementTypeTotalBodyWater:     true,
	MeasurementTypeLeanBodyMass:       true,
}

// IsValidMeasurementType returns true if the given measurement type is known.
func IsValidMeasurementType(mt MeasurementType) bool {
	return ValidMeasurementTypes[mt]
}

// ExtractionStatus represents whether extraction was full or partial.
type ExtractionStatus string

const (
	ExtractionStatusComplete ExtractionStatus = "complete"
	ExtractionStatusPartial  ExtractionStatus = "partial"
)

// Measurement represents a single body measurement value.
type Measurement struct {
	ID               string           `json:"id"`
	ClientID         string           `json:"client_id"`
	RecordedBy       string           `json:"recorded_by"`
	MeasurementType  MeasurementType  `json:"measurement_type"`
	Value            float64          `json:"value"`
	Unit             string           `json:"unit"`
	MeasuredAt       time.Time        `json:"measured_at"`
	Notes            string           `json:"notes,omitempty"`
	ArtifactID       string           `json:"artifact_id,omitempty"`
	ExtractionStatus ExtractionStatus `json:"extraction_status"`
	CreatedAt        time.Time        `json:"created_at"`
}

// ExtractionResult holds the output of an InBody PDF extraction.
type ExtractionResult struct {
	Measurements []*Measurement
	IsPartial    bool
	Errors       []string
}

// ExtractInBodyPayload is the JSON payload structure for extract_inbody jobs.
type ExtractInBodyPayload struct {
	ArtifactID string `json:"artifact_id"`
}

// ValidateNewMeasurement validates a measurement before creation.
func ValidateNewMeasurement(m *Measurement) error {
	if m.ClientID == "" {
		return NewValidationError("client_id is required")
	}
	if m.RecordedBy == "" {
		return NewValidationError("recorded_by is required")
	}
	if !IsValidMeasurementType(m.MeasurementType) {
		return NewValidationError("invalid measurement type: %s", m.MeasurementType)
	}
	if m.Unit == "" {
		return NewValidationError("unit is required")
	}
	if m.MeasuredAt.IsZero() {
		return NewValidationError("measured_at is required")
	}
	return nil
}
