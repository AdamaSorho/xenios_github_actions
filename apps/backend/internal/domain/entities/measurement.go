package entities

import "time"

// Measurement represents a single health measurement for a client.
type Measurement struct {
	ID              string   `json:"id"`
	ClientID        string   `json:"client_id"`
	RecordedBy      string   `json:"recorded_by"`
	MeasurementType string   `json:"type"`
	Value           float64  `json:"value"`
	Unit            string   `json:"unit"`
	MeasuredAt      time.Time `json:"measuredAt"`
	ArtifactID      *string  `json:"artifactId"`
	Flag            *string  `json:"flag"`
	ReferenceLow    *float64 `json:"referenceLow"`
	ReferenceHigh   *float64 `json:"referenceHigh"`
	Notes           *string  `json:"notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// MeasurementFilter holds query parameters for filtering measurements.
type MeasurementFilter struct {
	ClientID        string
	MeasurementType string
	From            *time.Time
	To              *time.Time
	Page            int
	Limit           int
}

// MeasurementResult holds a paginated list of measurements.
type MeasurementResult struct {
	Measurements []*Measurement `json:"measurements"`
	Pagination   Pagination     `json:"pagination"`
}

// Pagination holds pagination metadata.
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// LatestMeasurement represents the most recent measurement for a given type.
type LatestMeasurement struct {
	MeasurementType string   `json:"type"`
	Value           float64  `json:"value"`
	Unit            string   `json:"unit"`
	MeasuredAt      time.Time `json:"measuredAt"`
	Flag            *string  `json:"flag"`
	ReferenceLow    *float64 `json:"referenceLow"`
	ReferenceHigh   *float64 `json:"referenceHigh"`
}

// ValidMeasurementTypes enumerates known measurement types.
var ValidMeasurementTypes = map[string]bool{
	"weight":                true,
	"body_fat_pct":          true,
	"skeletal_muscle_mass":  true,
	"bmi":                   true,
	"waist_circumference":   true,
	"hip_circumference":     true,
	"resting_heart_rate":    true,
	"blood_pressure_sys":    true,
	"blood_pressure_dia":    true,
	"ldl_cholesterol":       true,
	"hdl_cholesterol":       true,
	"total_cholesterol":     true,
	"triglycerides":         true,
	"fasting_glucose":       true,
	"hba1c":                 true,
	"testosterone":          true,
	"cortisol":              true,
	"vitamin_d":             true,
	"iron":                  true,
	"creatinine":            true,
	"calories":              true,
	"protein":               true,
	"carbohydrates":         true,
	"fat":                   true,
}

// IsValidMeasurementType checks if a measurement type is known.
func IsValidMeasurementType(t string) bool {
	return ValidMeasurementTypes[t]
}

// ValidMeasurementFlags enumerates known measurement flags.
var ValidMeasurementFlags = map[string]bool{
	"low":      true,
	"high":     true,
	"critical": true,
	"normal":   true,
}

// IsValidMeasurementFlag checks if a flag value is known.
func IsValidMeasurementFlag(f string) bool {
	return ValidMeasurementFlags[f]
}
