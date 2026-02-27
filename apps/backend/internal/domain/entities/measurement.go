package entities

import "time"

// Measurement represents a health measurement (body composition, lab result, etc.)
type Measurement struct {
	ID           string     `json:"id"`
	ClientID     string     `json:"client_id"`
	Type         string     `json:"type"`
	Value        float64    `json:"value"`
	Unit         string     `json:"unit"`
	MeasuredAt   time.Time  `json:"measured_at"`
	ArtifactID   string     `json:"artifact_id,omitempty"`
	Flag         string     `json:"flag,omitempty"`
	ReferenceLow *float64   `json:"reference_low,omitempty"`
	ReferenceHigh *float64  `json:"reference_high,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// MeasurementFilter holds query parameters for filtering measurements.
type MeasurementFilter struct {
	ClientID string
	Type     string
	From     *time.Time
	To       *time.Time
	Page     int
	Limit    int
}

// MeasurementPage holds a paginated list of measurements.
type MeasurementPage struct {
	Measurements []*Measurement `json:"measurements"`
	Page         int            `json:"page"`
	Limit        int            `json:"limit"`
	Total        int            `json:"total"`
}

// LatestMeasurement represents the most recent measurement of a given type.
type LatestMeasurement struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
	Date  string  `json:"date"`
}

// ProfileSummary consolidates client health data for display.
type ProfileSummary struct {
	BodyComposition map[string]*LatestMeasurement `json:"bodyComposition"`
	Labs            *LabSummary                   `json:"labs"`
	Wearable        *WearableAverage              `json:"wearable"`
	Nutrition       *NutritionAverage             `json:"nutrition"`
}

// LabSummary holds flagged lab result information.
type LabSummary struct {
	FlaggedCount int                  `json:"flaggedCount"`
	LastTestDate string               `json:"lastTestDate"`
	Markers      []*LatestMeasurement `json:"markers"`
}

// WearableAverage holds rolling averages from wearable data.
type WearableAverage struct {
	Source        string   `json:"source"`
	AvgHRV7d     *float64 `json:"avgHrv7d"`
	AvgSleep7d   *float64 `json:"avgSleep7d"`
	AvgRecovery7d *float64 `json:"avgRecovery7d"`
}

// NutritionAverage holds rolling averages for nutrition data.
type NutritionAverage struct {
	AvgCalories7d *float64 `json:"avgCalories7d"`
	AvgProtein7d  *float64 `json:"avgProtein7d"`
}
