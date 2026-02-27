package entities

import "time"

// MeasurementFlag indicates whether a measurement value is out of range.
type MeasurementFlag string

const (
	MeasurementFlagNormal       MeasurementFlag = "normal"
	MeasurementFlagLow          MeasurementFlag = "low"
	MeasurementFlagHigh         MeasurementFlag = "high"
	MeasurementFlagCriticalLow  MeasurementFlag = "critical_low"
	MeasurementFlagCriticalHigh MeasurementFlag = "critical_high"
)

// Measurement represents a single health data point for a client.
type Measurement struct {
	ID              string          `json:"id"`
	ClientID        string          `json:"client_id"`
	RecordedBy      string          `json:"recorded_by"`
	MeasurementType string          `json:"measurement_type"`
	Value           float64         `json:"value"`
	Unit            string          `json:"unit"`
	Flag            MeasurementFlag `json:"flag"`
	RefRangeLow     *float64        `json:"ref_range_low,omitempty"`
	RefRangeHigh    *float64        `json:"ref_range_high,omitempty"`
	ArtifactID      string          `json:"artifact_id,omitempty"`
	MeasuredAt      time.Time       `json:"measured_at"`
	Notes           string          `json:"notes,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

// IsOutOfRange returns true if the measurement is flagged as out of range.
func (m *Measurement) IsOutOfRange() bool {
	switch m.Flag {
	case MeasurementFlagHigh, MeasurementFlagLow,
		MeasurementFlagCriticalHigh, MeasurementFlagCriticalLow:
		return true
	}
	return false
}

// IsCritical returns true if the measurement is critically out of range.
func (m *Measurement) IsCritical() bool {
	return m.Flag == MeasurementFlagCriticalHigh || m.Flag == MeasurementFlagCriticalLow
}
