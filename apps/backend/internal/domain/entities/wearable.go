package entities

import (
	"time"
)

// WearableSource represents a supported wearable device platform.
type WearableSource string

const (
	WearableSourceWhoop       WearableSource = "whoop"
	WearableSourceGarmin      WearableSource = "garmin"
	WearableSourceAppleHealth WearableSource = "apple_health"
	WearableSourceOura        WearableSource = "oura"
	WearableSourceFitbit      WearableSource = "fitbit"
)

// IsValidWearableSource returns true if the source is a known wearable platform.
func IsValidWearableSource(source WearableSource) bool {
	switch source {
	case WearableSourceWhoop,
		WearableSourceGarmin,
		WearableSourceAppleHealth,
		WearableSourceOura,
		WearableSourceFitbit:
		return true
	}
	return false
}

// MeasurementType represents the kind of measurement being recorded.
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

// IsValidWearableMeasurementType returns true if the type is a known wearable measurement.
func IsValidWearableMeasurementType(mt MeasurementType) bool {
	switch mt {
	case MeasurementTypeHRV,
		MeasurementTypeSleepDuration,
		MeasurementTypeRecovery,
		MeasurementTypeStrain,
		MeasurementTypeRestingHR,
		MeasurementTypeSteps,
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
		return "hours"
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

// WearableMeasurement represents a single normalized measurement from a wearable export.
type WearableMeasurement struct {
	ClientID        string          `json:"client_id"`
	Source          WearableSource  `json:"source"`
	MeasurementType MeasurementType `json:"measurement_type"`
	Value           float64         `json:"value"`
	Unit            string          `json:"unit"`
	MeasuredAt      time.Time       `json:"measured_at"`
}

// WearableSummary holds rolling averages for a client from a wearable source.
type WearableSummary struct {
	ID          string                 `json:"id"`
	ClientID    string                 `json:"client_id"`
	Source      WearableSource         `json:"source"`
	SummaryDate time.Time              `json:"summary_date"`
	Metrics     map[string]interface{} `json:"metrics"`
	SyncedAt    time.Time              `json:"synced_at"`
	CreatedAt   time.Time              `json:"created_at"`
}

// RollingAverageWindows defines the standard windows for rolling average computation.
var RollingAverageWindows = []int{7, 14, 30}

// SourceMetricSupport maps each wearable source to the metrics it supports.
var SourceMetricSupport = map[WearableSource][]MeasurementType{
	WearableSourceWhoop: {
		MeasurementTypeHRV,
		MeasurementTypeSleepDuration,
		MeasurementTypeRecovery,
		MeasurementTypeStrain,
		MeasurementTypeRestingHR,
	},
	WearableSourceGarmin: {
		MeasurementTypeHRV,
		MeasurementTypeSleepDuration,
		MeasurementTypeSteps,
		MeasurementTypeRestingHR,
	},
	WearableSourceAppleHealth: {
		MeasurementTypeHRV,
		MeasurementTypeSleepDuration,
		MeasurementTypeSteps,
		MeasurementTypeRestingHR,
	},
	WearableSourceOura: {
		MeasurementTypeHRV,
		MeasurementTypeSleepDuration,
		MeasurementTypeRecovery,
		MeasurementTypeSteps,
		MeasurementTypeRestingHR,
	},
	WearableSourceFitbit: {
		MeasurementTypeHRV,
		MeasurementTypeSleepDuration,
		MeasurementTypeSteps,
		MeasurementTypeRestingHR,
	},
}
