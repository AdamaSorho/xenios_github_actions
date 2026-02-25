package entities

import "time"

// Measurement represents a single health measurement (body composition, lab result, etc.).
type Measurement struct {
	ID              string   `json:"id"`
	ClientID        string   `json:"client_id"`
	Type            string   `json:"type"`
	Value           float64  `json:"value"`
	Unit            string   `json:"unit"`
	MeasuredAt      time.Time `json:"measured_at"`
	ArtifactID      *string  `json:"artifact_id,omitempty"`
	Flag            *string  `json:"flag,omitempty"`
	ReferenceLow    *float64 `json:"reference_low,omitempty"`
	ReferenceHigh   *float64 `json:"reference_high,omitempty"`
	Notes           *string  `json:"notes,omitempty"`
	RecordedBy      string   `json:"recorded_by"`
	CreatedAt       time.Time `json:"created_at"`
}

// MeasurementFilter holds query parameters for listing measurements.
type MeasurementFilter struct {
	ClientID string
	Type     string
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
}
