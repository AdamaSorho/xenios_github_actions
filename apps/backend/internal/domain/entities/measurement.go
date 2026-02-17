package entities

import "time"

// Measurement represents a single health measurement for a client.
type Measurement struct {
	ID              string    `json:"id"`
	ClientID        string    `json:"client_id"`
	RecordedBy      string    `json:"recorded_by"`
	MeasurementType string    `json:"type"`
	Value           float64   `json:"value"`
	Unit            string    `json:"unit"`
	MeasuredAt      time.Time `json:"measured_at"`
	Notes           *string   `json:"notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// MeasurementFilter holds query parameters for listing measurements.
type MeasurementFilter struct {
	ClientID        string
	MeasurementType string
	From            *time.Time
	To              *time.Time
	Limit           int
	Offset          int
}

// LatestMeasurement represents the most recent value for a measurement type.
type LatestMeasurement struct {
	MeasurementType string    `json:"type"`
	Value           float64   `json:"value"`
	Unit            string    `json:"unit"`
	MeasuredAt      time.Time `json:"measured_at"`
}
