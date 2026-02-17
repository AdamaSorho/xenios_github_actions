package entities

import "time"

// MeasurementFlag indicates whether a value is within normal reference range.
type MeasurementFlag string

const (
	FlagNormal       MeasurementFlag = "normal"
	FlagLow          MeasurementFlag = "low"
	FlagHigh         MeasurementFlag = "high"
	FlagCriticalLow  MeasurementFlag = "critical_low"
	FlagCriticalHigh MeasurementFlag = "critical_high"
)

// IsValidMeasurementFlag returns true if the flag is a known value.
func IsValidMeasurementFlag(f MeasurementFlag) bool {
	switch f {
	case FlagNormal, FlagLow, FlagHigh, FlagCriticalLow, FlagCriticalHigh:
		return true
	}
	return false
}

// LabMeasurementType represents a recognized lab biomarker type.
type LabMeasurementType string

const (
	LabFastingGlucose   LabMeasurementType = "fasting_glucose"
	LabTotalCholesterol LabMeasurementType = "total_cholesterol"
	LabLDLCholesterol   LabMeasurementType = "ldl_cholesterol"
	LabHDLCholesterol   LabMeasurementType = "hdl_cholesterol"
	LabTriglycerides    LabMeasurementType = "triglycerides"
	LabHbA1c            LabMeasurementType = "hba1c"
	LabTestosterone     LabMeasurementType = "testosterone"
	LabTSH              LabMeasurementType = "tsh"
	LabVitaminD         LabMeasurementType = "vitamin_d"
	LabIron             LabMeasurementType = "iron"
)

// KnownLabMarkers maps common lab result names (lowercased) to their measurement type.
var KnownLabMarkers = map[string]LabMeasurementType{
	"glucose":               LabFastingGlucose,
	"glucose, fasting":      LabFastingGlucose,
	"fasting glucose":       LabFastingGlucose,
	"glucose fasting":       LabFastingGlucose,
	"total cholesterol":     LabTotalCholesterol,
	"cholesterol, total":    LabTotalCholesterol,
	"cholesterol total":     LabTotalCholesterol,
	"cholesterol":           LabTotalCholesterol,
	"ldl cholesterol":       LabLDLCholesterol,
	"ldl-c":                 LabLDLCholesterol,
	"ldl":                   LabLDLCholesterol,
	"ldl cholesterol, calc": LabLDLCholesterol,
	"hdl cholesterol":       LabHDLCholesterol,
	"hdl-c":                 LabHDLCholesterol,
	"hdl":                   LabHDLCholesterol,
	"triglycerides":         LabTriglycerides,
	"triglyceride":          LabTriglycerides,
	"hba1c":                 LabHbA1c,
	"hemoglobin a1c":        LabHbA1c,
	"a1c":                   LabHbA1c,
	"testosterone":          LabTestosterone,
	"testosterone, total":   LabTestosterone,
	"total testosterone":    LabTestosterone,
	"tsh":                   LabTSH,
	"thyroid stimulating hormone": LabTSH,
	"vitamin d":                   LabVitaminD,
	"vitamin d, 25-hydroxy":       LabVitaminD,
	"25-hydroxyvitamin d":         LabVitaminD,
	"iron":                        LabIron,
	"iron, serum":                 LabIron,
	"serum iron":                  LabIron,
}

// Measurement represents a single lab result measurement.
type Measurement struct {
	ID              string          `json:"id"`
	ClientID        string          `json:"client_id"`
	RecordedBy      string          `json:"recorded_by"`
	MeasurementType string          `json:"measurement_type"`
	Value           float64         `json:"value"`
	Unit            string          `json:"unit"`
	MeasuredAt      time.Time       `json:"measured_at"`
	Notes           string          `json:"notes,omitempty"`
	ReferenceLow    *float64        `json:"reference_low,omitempty"`
	ReferenceHigh   *float64        `json:"reference_high,omitempty"`
	Flag            MeasurementFlag `json:"flag,omitempty"`
	ArtifactID      *string         `json:"artifact_id,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

// LabResult is a parsed biomarker from a lab report before it becomes a Measurement.
type LabResult struct {
	MarkerName   string
	Value        float64
	Unit         string
	ReferenceLow *float64
	ReferenceHigh *float64
	Flag         MeasurementFlag
	RawName      string
}

// DetermineFlag computes the flag based on value and reference range.
// Returns empty string if no reference range is available.
func DetermineFlag(value float64, refLow, refHigh *float64) MeasurementFlag {
	if refLow == nil && refHigh == nil {
		return ""
	}

	if refLow != nil && refHigh != nil {
		if value < *refLow {
			return FlagLow
		}
		if value > *refHigh {
			return FlagHigh
		}
		return FlagNormal
	}

	if refLow != nil && refHigh == nil {
		// Only lower bound (e.g., HDL > 40)
		if value < *refLow {
			return FlagLow
		}
		return FlagNormal
	}

	// Only upper bound (e.g., LDL < 100)
	if value > *refHigh {
		return FlagHigh
	}
	return FlagNormal
}
