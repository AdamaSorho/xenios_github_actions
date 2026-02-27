package entities

import (
	"math"
	"strings"
)

// LabMeasurementType represents a normalized lab marker type.
type LabMeasurementType string

const (
	LabTypeFastingGlucose   LabMeasurementType = "fasting_glucose"
	LabTypeTotalCholesterol  LabMeasurementType = "total_cholesterol"
	LabTypeLDLCholesterol    LabMeasurementType = "ldl_cholesterol"
	LabTypeHDLCholesterol    LabMeasurementType = "hdl_cholesterol"
	LabTypeTriglycerides     LabMeasurementType = "triglycerides"
	LabTypeHbA1c             LabMeasurementType = "hba1c"
	LabTypeTestosterone      LabMeasurementType = "testosterone"
	LabTypeTSH               LabMeasurementType = "tsh"
	LabTypeVitaminD          LabMeasurementType = "vitamin_d"
	LabTypeIron              LabMeasurementType = "iron"
)

// LabFlag represents the out-of-range status of a measurement.
type LabFlag string

const (
	LabFlagNormal      LabFlag = "normal"
	LabFlagLow         LabFlag = "low"
	LabFlagHigh        LabFlag = "high"
	LabFlagCriticalLow LabFlag = "critical_low"
	LabFlagCriticalHigh LabFlag = "critical_high"
)

// LabMeasurement represents a single extracted lab marker with reference range.
type LabMeasurement struct {
	MeasurementType LabMeasurementType `json:"measurement_type"`
	Value           float64            `json:"value"`
	Unit            string             `json:"unit"`
	ReferenceLow    *float64           `json:"reference_low,omitempty"`
	ReferenceHigh   *float64           `json:"reference_high,omitempty"`
	Flag            *LabFlag           `json:"flag,omitempty"`
}

// LabResultPayload is the job payload for extract_lab_results jobs.
type LabResultPayload struct {
	ArtifactID string `json:"artifact_id"`
	CoachID    string `json:"coach_id"`
	ClientID   string `json:"client_id"`
}

// markerAliases maps common lab report names to normalized types.
var markerAliases = map[string]LabMeasurementType{
	"glucose":           LabTypeFastingGlucose,
	"glucose, fasting":  LabTypeFastingGlucose,
	"fasting glucose":   LabTypeFastingGlucose,
	"glucose fasting":   LabTypeFastingGlucose,
	"total cholesterol": LabTypeTotalCholesterol,
	"cholesterol":       LabTypeTotalCholesterol,
	"cholesterol total": LabTypeTotalCholesterol,
	"cholesterol, total": LabTypeTotalCholesterol,
	"ldl cholesterol":   LabTypeLDLCholesterol,
	"ldl-c":             LabTypeLDLCholesterol,
	"ldl":               LabTypeLDLCholesterol,
	"ldl chol calc":     LabTypeLDLCholesterol,
	"hdl cholesterol":   LabTypeHDLCholesterol,
	"hdl-c":             LabTypeHDLCholesterol,
	"hdl":               LabTypeHDLCholesterol,
	"triglycerides":     LabTypeTriglycerides,
	"triglyceride":      LabTypeTriglycerides,
	"hba1c":             LabTypeHbA1c,
	"hemoglobin a1c":    LabTypeHbA1c,
	"a1c":               LabTypeHbA1c,
	"testosterone":      LabTypeTestosterone,
	"testosterone, total": LabTypeTestosterone,
	"total testosterone":  LabTypeTestosterone,
	"tsh":               LabTypeTSH,
	"thyroid stimulating hormone": LabTypeTSH,
	"vitamin d":         LabTypeVitaminD,
	"vitamin d, 25-hydroxy": LabTypeVitaminD,
	"25-hydroxy vitamin d":  LabTypeVitaminD,
	"iron":              LabTypeIron,
	"iron, serum":       LabTypeIron,
	"serum iron":        LabTypeIron,
}

// NormalizeMarkerName maps a lab report marker name to a LabMeasurementType.
// Returns empty string if the marker is not recognized.
func NormalizeMarkerName(name string) LabMeasurementType {
	normalized := strings.ToLower(strings.TrimSpace(name))
	// Remove extra spaces
	parts := strings.Fields(normalized)
	normalized = strings.Join(parts, " ")

	if mt, ok := markerAliases[normalized]; ok {
		return mt
	}
	return ""
}

// DetermineFlag evaluates a value against reference ranges and returns the flag.
// Returns nil if no reference range is provided.
func DetermineFlag(value float64, referenceLow, referenceHigh *float64) *LabFlag {
	if referenceLow == nil && referenceHigh == nil {
		return nil
	}

	flag := LabFlagNormal

	if referenceLow != nil && value < *referenceLow {
		flag = LabFlagLow
		// Critical if more than 50% below reference low
		if *referenceLow > 0 {
			threshold := *referenceLow * 0.5
			if value < threshold {
				flag = LabFlagCriticalLow
			}
		}
	}

	if referenceHigh != nil && value > *referenceHigh {
		flag = LabFlagHigh
		// Critical if more than 50% above reference high
		if *referenceHigh > 0 {
			threshold := *referenceHigh * 1.5
			if value > threshold {
				flag = LabFlagCriticalHigh
			}
		}
	}

	return &flag
}

// IsValidLabMeasurementType returns true if the type is a known lab measurement type.
func IsValidLabMeasurementType(mt LabMeasurementType) bool {
	switch mt {
	case LabTypeFastingGlucose,
		LabTypeTotalCholesterol,
		LabTypeLDLCholesterol,
		LabTypeHDLCholesterol,
		LabTypeTriglycerides,
		LabTypeHbA1c,
		LabTypeTestosterone,
		LabTypeTSH,
		LabTypeVitaminD,
		LabTypeIron:
		return true
	}
	return false
}

// floatPtr is a helper to create a float64 pointer.
func FloatPtr(f float64) *float64 {
	return &f
}

// RoundToThreeDecimals rounds a float to 3 decimal places for storage in NUMERIC(10,3).
func RoundToThreeDecimals(f float64) float64 {
	return math.Round(f*1000) / 1000
}
