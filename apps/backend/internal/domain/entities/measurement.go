package entities

import "time"

// MeasurementType represents the type of body composition measurement.
type MeasurementType string

const (
	MeasurementTypeWeight             MeasurementType = "weight"
	MeasurementTypeBodyFatPct         MeasurementType = "body_fat_pct"
	MeasurementTypeSkeletalMuscleMass MeasurementType = "skeletal_muscle_mass"
	MeasurementTypeBMR                MeasurementType = "bmr"
	MeasurementTypeTotalBodyWater     MeasurementType = "total_body_water"
	MeasurementTypeLeanBodyMass       MeasurementType = "lean_body_mass"
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

// IsValidMeasurementType checks if a measurement type is recognized.
func IsValidMeasurementType(mt MeasurementType) bool {
	return ValidMeasurementTypes[mt]
}

// Measurement represents a body composition measurement record.
type Measurement struct {
	ID                string          `json:"id"`
	ClientID          string          `json:"client_id"`
	RecordedBy        string          `json:"recorded_by"`
	MeasurementType   MeasurementType `json:"measurement_type"`
	Value             float64         `json:"value"`
	Unit              string          `json:"unit"`
	MeasuredAt        time.Time       `json:"measured_at"`
	Notes             string          `json:"notes,omitempty"`
	ArtifactID        string          `json:"artifact_id,omitempty"`
	PartialExtraction bool            `json:"partial_extraction"`
	CreatedAt         time.Time       `json:"created_at"`
}

// InBodyResult holds the structured extraction result from an InBody PDF.
type InBodyResult struct {
	Weight             *float64 `json:"weight,omitempty"`
	WeightUnit         string   `json:"weight_unit,omitempty"`
	BodyFatPct         *float64 `json:"body_fat_pct,omitempty"`
	SkeletalMuscleMass *float64 `json:"skeletal_muscle_mass,omitempty"`
	SMMUnit            string   `json:"smm_unit,omitempty"`
	BMR                *float64 `json:"bmr,omitempty"`
	TotalBodyWater     *float64 `json:"total_body_water,omitempty"`
	LeanBodyMass       *float64 `json:"lean_body_mass,omitempty"`
	LBMUnit            string   `json:"lbm_unit,omitempty"`
	MeasuredAt         *time.Time `json:"measured_at,omitempty"`
}

// ExtractedFieldCount returns the number of successfully extracted fields.
func (r *InBodyResult) ExtractedFieldCount() int {
	count := 0
	if r.Weight != nil {
		count++
	}
	if r.BodyFatPct != nil {
		count++
	}
	if r.SkeletalMuscleMass != nil {
		count++
	}
	if r.BMR != nil {
		count++
	}
	if r.TotalBodyWater != nil {
		count++
	}
	if r.LeanBodyMass != nil {
		count++
	}
	return count
}

// TotalExpectedFields is the number of fields we attempt to extract.
const TotalExpectedFields = 6

// IsPartial returns true if not all expected fields were extracted.
func (r *InBodyResult) IsPartial() bool {
	return r.ExtractedFieldCount() < TotalExpectedFields
}

// IsEmpty returns true if no fields were extracted at all.
func (r *InBodyResult) IsEmpty() bool {
	return r.ExtractedFieldCount() == 0
}
