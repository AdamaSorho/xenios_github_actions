package entities

import (
	"time"
)

// WearableSource identifies the wearable device platform.
type WearableSource string

const (
	WearableSourceWhoop       WearableSource = "whoop"
	WearableSourceGarmin      WearableSource = "garmin"
	WearableSourceAppleHealth WearableSource = "apple_health"
	WearableSourceOura        WearableSource = "oura"
	WearableSourceFitbit      WearableSource = "fitbit"
)

// IsValidWearableSource returns true if the given source is recognized.
func IsValidWearableSource(s WearableSource) bool {
	switch s {
	case WearableSourceWhoop, WearableSourceGarmin, WearableSourceAppleHealth,
		WearableSourceOura, WearableSourceFitbit:
		return true
	}
	return false
}

// MeasurementType represents a specific type of health measurement.
type MeasurementType string

const (
	MeasurementTypeHRV           MeasurementType = "hrv_ms"
	MeasurementTypeSleepDuration MeasurementType = "sleep_duration_hrs"
	MeasurementTypeRecovery      MeasurementType = "recovery_score"
	MeasurementTypeStrain        MeasurementType = "strain_score"
	MeasurementTypeRestingHR     MeasurementType = "resting_hr_bpm"
	MeasurementTypeSteps         MeasurementType = "steps_count"
	MeasurementTypeSleepQuality  MeasurementType = "sleep_quality_score"
)

// IsValidMeasurementType returns true if the measurement type is recognized.
func IsValidMeasurementType(mt MeasurementType) bool {
	switch mt {
	case MeasurementTypeHRV, MeasurementTypeSleepDuration, MeasurementTypeRecovery,
		MeasurementTypeStrain, MeasurementTypeRestingHR, MeasurementTypeSteps,
		MeasurementTypeSleepQuality:
		return true
	}
	return false
}

// UnitForMeasurementType returns the standard unit for a measurement type.
func UnitForMeasurementType(mt MeasurementType) string {
	switch mt {
	case MeasurementTypeHRV:
		return "ms"
	case MeasurementTypeSleepDuration:
		return "hrs"
	case MeasurementTypeRecovery:
		return "score"
	case MeasurementTypeStrain:
		return "score"
	case MeasurementTypeRestingHR:
		return "bpm"
	case MeasurementTypeSteps:
		return "count"
	case MeasurementTypeSleepQuality:
		return "score"
	default:
		return ""
	}
}

// Measurement represents a single health metric data point.
type Measurement struct {
	ID              string          `json:"id"`
	ClientID        string          `json:"client_id"`
	RecordedBy      string          `json:"recorded_by"`
	MeasurementType MeasurementType `json:"measurement_type"`
	Value           float64         `json:"value"`
	Unit            string          `json:"unit"`
	MeasuredAt      time.Time       `json:"measured_at"`
	Notes           string          `json:"notes,omitempty"`
	Source          WearableSource  `json:"source"`
	CreatedAt       time.Time       `json:"created_at"`
}

// WearableSummary contains aggregated wearable metrics for a client.
type WearableSummary struct {
	ID          string                 `json:"id"`
	ClientID    string                 `json:"client_id"`
	Source      WearableSource         `json:"source"`
	SummaryDate time.Time              `json:"summary_date"`
	Metrics     map[string]interface{} `json:"metrics"`
	SyncedAt    time.Time              `json:"synced_at"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ExtractWearablePayload is the payload for an extract_wearable job.
type ExtractWearablePayload struct {
	ArtifactID string         `json:"artifact_id"`
	ClientID   string         `json:"client_id"`
	CoachID    string         `json:"coach_id"`
	Source     WearableSource `json:"source"`
}

// RollingAverageWindows defines the window sizes for rolling average computation.
var RollingAverageWindows = []int{7, 14, 30}
