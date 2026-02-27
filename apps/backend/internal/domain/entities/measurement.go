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
var ValidMeasurementTypes = []MeasurementType{
	MeasurementTypeWeight,
	MeasurementTypeBodyFatPct,
	MeasurementTypeSkeletalMuscleMass,
	MeasurementTypeBMR,
	MeasurementTypeTotalBodyWater,
	MeasurementTypeLeanBodyMass,
}

// IsValidMeasurementType returns true if the given type is recognized.
func IsValidMeasurementType(mt MeasurementType) bool {
	for _, valid := range ValidMeasurementTypes {
		if mt == valid {
			return true
		}
	}
	return false
}

// Measurement represents a single body composition measurement record.
type Measurement struct {
	ID              string          `json:"id"`
	ClientID        string          `json:"client_id"`
	RecordedBy      string          `json:"recorded_by"`
	MeasurementType MeasurementType `json:"measurement_type"`
	Value           float64         `json:"value"`
	Unit            string          `json:"unit"`
	MeasuredAt      time.Time       `json:"measured_at"`
	Notes           *string         `json:"notes,omitempty"`
	ArtifactID      *string         `json:"artifact_id,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

// ExtractedMetric represents a single metric extracted from an InBody PDF.
type ExtractedMetric struct {
	Type  MeasurementType
	Value float64
	Unit  string
}

// InBodyResult holds the result of parsing InBody text.
type InBodyResult struct {
	Metrics    []ExtractedMetric
	Partial    bool
	MeasuredAt *time.Time
}
