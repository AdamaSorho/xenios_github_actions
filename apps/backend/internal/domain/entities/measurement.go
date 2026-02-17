package entities

import "time"

// MeasurementType represents the category of a health measurement.
type MeasurementType string

const (
	MeasurementTypeLab         MeasurementType = "lab"
	MeasurementTypeWearable    MeasurementType = "wearable"
	MeasurementTypeBodyComp    MeasurementType = "body_composition"
)

// MeasurementFlag indicates whether a lab value is normal, high, low, or critical.
type MeasurementFlag string

const (
	MeasurementFlagNormal       MeasurementFlag = "normal"
	MeasurementFlagHigh         MeasurementFlag = "high"
	MeasurementFlagLow          MeasurementFlag = "low"
	MeasurementFlagCriticalHigh MeasurementFlag = "critical_high"
	MeasurementFlagCriticalLow  MeasurementFlag = "critical_low"
)

// Measurement represents a single health data point extracted from an artifact.
type Measurement struct {
	ID           string          `json:"id"`
	ClientID     string          `json:"client_id"`
	CoachID      string          `json:"coach_id"`
	ArtifactID   string          `json:"artifact_id"`
	Type         MeasurementType `json:"type"`
	MarkerName   string          `json:"marker_name"`
	Value        float64         `json:"value"`
	Unit         string          `json:"unit"`
	ReferenceMin *float64        `json:"reference_min,omitempty"`
	ReferenceMax *float64        `json:"reference_max,omitempty"`
	Flag         MeasurementFlag `json:"flag"`
	RecordedAt   time.Time       `json:"recorded_at"`
	CreatedAt    time.Time       `json:"created_at"`
}

// IsOutOfRange returns true if the measurement is flagged high or low.
func (m *Measurement) IsOutOfRange() bool {
	switch m.Flag {
	case MeasurementFlagHigh, MeasurementFlagLow:
		return true
	}
	return false
}

// IsCritical returns true if the measurement is flagged as critical.
func (m *Measurement) IsCritical() bool {
	switch m.Flag {
	case MeasurementFlagCriticalHigh, MeasurementFlagCriticalLow:
		return true
	}
	return false
}
