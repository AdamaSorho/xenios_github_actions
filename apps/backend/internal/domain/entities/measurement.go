package entities

import "time"

// MeasurementFlag indicates whether a measurement is within normal range.
type MeasurementFlag string

const (
	MeasurementFlagNormal       MeasurementFlag = "normal"
	MeasurementFlagLow          MeasurementFlag = "low"
	MeasurementFlagHigh         MeasurementFlag = "high"
	MeasurementFlagCriticalLow  MeasurementFlag = "critical_low"
	MeasurementFlagCriticalHigh MeasurementFlag = "critical_high"
)

// Measurement represents a health data point recorded for a client.
type Measurement struct {
	ID              string          `json:"id"`
	ClientID        string          `json:"client_id"`
	RecordedBy      string          `json:"recorded_by"`
	MeasurementType string          `json:"measurement_type"`
	Value           float64         `json:"value"`
	Unit            string          `json:"unit"`
	Flag            MeasurementFlag `json:"flag,omitempty"`
	MeasuredAt      time.Time       `json:"measured_at"`
	Notes           string          `json:"notes,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

// ReferenceRange defines the acceptable range for a measurement type.
type ReferenceRange struct {
	MeasurementType string  `json:"measurement_type"`
	Unit            string  `json:"unit"`
	LowCritical     float64 `json:"low_critical"`
	LowNormal       float64 `json:"low_normal"`
	HighNormal      float64 `json:"high_normal"`
	HighCritical    float64 `json:"high_critical"`
	DisplayName     string  `json:"display_name"`
}

// EvaluateFlag determines the flag for a measurement value against the reference range.
func (r *ReferenceRange) EvaluateFlag(value float64) MeasurementFlag {
	if value <= r.LowCritical {
		return MeasurementFlagCriticalLow
	}
	if value < r.LowNormal {
		return MeasurementFlagLow
	}
	if value > r.HighCritical {
		return MeasurementFlagCriticalHigh
	}
	if value > r.HighNormal {
		return MeasurementFlagHigh
	}
	return MeasurementFlagNormal
}

// LabReferenceRanges maps measurement types to their reference ranges.
// These are standard clinical reference ranges for common lab markers.
var LabReferenceRanges = map[string]*ReferenceRange{
	"ldl_cholesterol": {
		MeasurementType: "ldl_cholesterol",
		Unit:            "mg/dL",
		LowCritical:     0,
		LowNormal:       0,
		HighNormal:      100,
		HighCritical:    190,
		DisplayName:     "LDL Cholesterol",
	},
	"hdl_cholesterol": {
		MeasurementType: "hdl_cholesterol",
		Unit:            "mg/dL",
		LowCritical:     20,
		LowNormal:       40,
		HighNormal:      100,
		HighCritical:    100,
		DisplayName:     "HDL Cholesterol",
	},
	"fasting_glucose": {
		MeasurementType: "fasting_glucose",
		Unit:            "mg/dL",
		LowCritical:     40,
		LowNormal:       70,
		HighNormal:      100,
		HighCritical:    200,
		DisplayName:     "Fasting Glucose",
	},
	"total_cholesterol": {
		MeasurementType: "total_cholesterol",
		Unit:            "mg/dL",
		LowCritical:     0,
		LowNormal:       0,
		HighNormal:      200,
		HighCritical:    300,
		DisplayName:     "Total Cholesterol",
	},
	"triglycerides": {
		MeasurementType: "triglycerides",
		Unit:            "mg/dL",
		LowCritical:     0,
		LowNormal:       0,
		HighNormal:      150,
		HighCritical:    500,
		DisplayName:     "Triglycerides",
	},
	"hemoglobin_a1c": {
		MeasurementType: "hemoglobin_a1c",
		Unit:            "%",
		LowCritical:     0,
		LowNormal:       0,
		HighNormal:      5.7,
		HighCritical:    9.0,
		DisplayName:     "Hemoglobin A1c",
	},
}
