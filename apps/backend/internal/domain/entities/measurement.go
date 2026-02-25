package entities

import "time"

// WearableSource identifies the wearable platform the data originated from.
type WearableSource string

const (
	WearableSourceWhoop       WearableSource = "whoop"
	WearableSourceGarmin      WearableSource = "garmin"
	WearableSourceAppleHealth WearableSource = "apple_health"
	WearableSourceOura        WearableSource = "oura"
	WearableSourceFitbit      WearableSource = "fitbit"
)

// IsValidWearableSource returns true if the source is one of the known wearable platforms.
func IsValidWearableSource(s WearableSource) bool {
	switch s {
	case WearableSourceWhoop, WearableSourceGarmin, WearableSourceAppleHealth,
		WearableSourceOura, WearableSourceFitbit:
		return true
	}
	return false
}

// MeasurementType categorises the kind of metric captured.
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

// IsValidMeasurementType returns true if the measurement type is known.
func IsValidMeasurementType(mt MeasurementType) bool {
	switch mt {
	case MeasurementTypeHRV, MeasurementTypeSleepDuration, MeasurementTypeRecovery,
		MeasurementTypeStrain, MeasurementTypeRestingHR, MeasurementTypeSteps,
		MeasurementTypeSleepQuality:
		return true
	}
	return false
}

// Measurement represents a single daily metric reading from a wearable device.
type Measurement struct {
	ID              string          `json:"id"`
	ClientID        string          `json:"client_id"`
	Source          WearableSource  `json:"source"`
	MeasurementType MeasurementType `json:"measurement_type"`
	Value           float64         `json:"value"`
	MeasuredAt      time.Time       `json:"measured_at"`
	CreatedAt       time.Time       `json:"created_at"`
}
