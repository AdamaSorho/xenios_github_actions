package entities

import "time"

// MeasurementType represents the kind of measurement being recorded.
type MeasurementType string

const (
	MeasurementTypeWeight            MeasurementType = "weight"
	MeasurementTypeBodyFatPct        MeasurementType = "body_fat_pct"
	MeasurementTypeSkeletalMuscleMass MeasurementType = "skeletal_muscle_mass"
	MeasurementTypeBMR               MeasurementType = "bmr"
	MeasurementTypeTotalBodyWater    MeasurementType = "total_body_water"
	MeasurementTypeLeanBodyMass      MeasurementType = "lean_body_mass"
)

// ValidMeasurementTypes lists all known measurement types.
var ValidMeasurementTypes = map[MeasurementType]bool{
	MeasurementTypeWeight:             true,
	MeasurementTypeBodyFatPct:         true,
	MeasurementTypeSkeletalMuscleMass: true,
	MeasurementTypeBMR:                true,
	MeasurementTypeTotalBodyWater:     true,
	MeasurementTypeLeanBodyMass:       true,
}

// IsValidMeasurementType checks if a measurement type is known.
func IsValidMeasurementType(mt MeasurementType) bool {
	return ValidMeasurementTypes[mt]
}

// Measurement represents a single health measurement for a client.
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

// ValidateMeasurement validates a measurement's required fields.
func ValidateMeasurement(m *Measurement) error {
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

// InBodyResult holds the structured output of an InBody PDF extraction.
type InBodyResult struct {
	Weight            *float64 `json:"weight,omitempty"`
	WeightUnit        string   `json:"weight_unit,omitempty"`
	BodyFatPct        *float64 `json:"body_fat_pct,omitempty"`
	SkeletalMuscleMass *float64 `json:"skeletal_muscle_mass,omitempty"`
	SkeletalMuscleUnit string   `json:"skeletal_muscle_unit,omitempty"`
	BMR               *float64 `json:"bmr,omitempty"`
	TotalBodyWater    *float64 `json:"total_body_water,omitempty"`
	LeanBodyMass      *float64 `json:"lean_body_mass,omitempty"`
	LeanBodyMassUnit  string   `json:"lean_body_mass_unit,omitempty"`
	MeasuredAt        time.Time `json:"measured_at"`
	IsPartial         bool     `json:"is_partial"`
}

// FieldCount returns the number of successfully extracted fields.
func (r *InBodyResult) FieldCount() int {
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

// TotalFields is the total number of fields the extractor attempts to extract.
const TotalFields = 6

// ToMeasurements converts an InBodyResult into a slice of Measurement entities.
func (r *InBodyResult) ToMeasurements(clientID, recordedBy, artifactID string) []*Measurement {
	var measurements []*Measurement

	add := func(mt MeasurementType, value *float64, unit string) {
		if value == nil {
			return
		}
		measurements = append(measurements, &Measurement{
			ClientID:        clientID,
			RecordedBy:      recordedBy,
			MeasurementType: mt,
			Value:           *value,
			Unit:            unit,
			MeasuredAt:      r.MeasuredAt,
			ArtifactID:      artifactID,
		})
	}

	add(MeasurementTypeWeight, r.Weight, r.WeightUnit)
	add(MeasurementTypeBodyFatPct, r.BodyFatPct, "%")
	add(MeasurementTypeSkeletalMuscleMass, r.SkeletalMuscleMass, r.SkeletalMuscleUnit)
	add(MeasurementTypeBMR, r.BMR, "kcal")
	add(MeasurementTypeTotalBodyWater, r.TotalBodyWater, "L")
	add(MeasurementTypeLeanBodyMass, r.LeanBodyMass, r.LeanBodyMassUnit)

	return measurements
}
