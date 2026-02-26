package entities

import "time"

// MeasurementFlag represents the out-of-range status of a measurement.
type MeasurementFlag string

const (
	FlagNormal       MeasurementFlag = "normal"
	FlagLow          MeasurementFlag = "low"
	FlagHigh         MeasurementFlag = "high"
	FlagCriticalLow  MeasurementFlag = "critical_low"
	FlagCriticalHigh MeasurementFlag = "critical_high"
)

// IsValidMeasurementFlag returns true if the flag is a recognized value.
func IsValidMeasurementFlag(f MeasurementFlag) bool {
	switch f {
	case FlagNormal, FlagLow, FlagHigh, FlagCriticalLow, FlagCriticalHigh:
		return true
	}
	return false
}

// Measurement represents a single health measurement with optional reference range.
type Measurement struct {
	ID              string           `json:"id"`
	ClientID        string           `json:"client_id"`
	RecordedBy      string           `json:"recorded_by"`
	MeasurementType string           `json:"measurement_type"`
	Value           float64          `json:"value"`
	Unit            string           `json:"unit"`
	MeasuredAt      time.Time        `json:"measured_at"`
	Notes           *string          `json:"notes,omitempty"`
	ArtifactID      *string          `json:"artifact_id,omitempty"`
	ReferenceLow    *float64         `json:"reference_low,omitempty"`
	ReferenceHigh   *float64         `json:"reference_high,omitempty"`
	Flag            *MeasurementFlag `json:"flag,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
}

// ParsedMarker is a raw marker extracted from a lab file before normalization.
type ParsedMarker struct {
	Name          string
	Value         float64
	Unit          string
	ReferenceLow  *float64
	ReferenceHigh *float64
}

// labMarkerAliases maps various marker name variants to canonical measurement types.
var labMarkerAliases = map[string]string{
	"glucose":                "fasting_glucose",
	"fasting glucose":        "fasting_glucose",
	"glucose, fasting":       "fasting_glucose",
	"total cholesterol":      "total_cholesterol",
	"cholesterol":            "total_cholesterol",
	"cholesterol, total":     "total_cholesterol",
	"ldl":                    "ldl_cholesterol",
	"ldl cholesterol":        "ldl_cholesterol",
	"ldl-c":                  "ldl_cholesterol",
	"ldl chol calc":          "ldl_cholesterol",
	"hdl":                    "hdl_cholesterol",
	"hdl cholesterol":        "hdl_cholesterol",
	"hdl-c":                  "hdl_cholesterol",
	"triglycerides":          "triglycerides",
	"triglyceride":           "triglycerides",
	"hba1c":                  "hba1c",
	"hemoglobin a1c":         "hba1c",
	"a1c":                    "hba1c",
	"testosterone":           "testosterone",
	"testosterone, total":    "testosterone",
	"tsh":                    "tsh",
	"thyroid stimulating hormone": "tsh",
	"vitamin d":              "vitamin_d",
	"vitamin d, 25-hydroxy":  "vitamin_d",
	"25-hydroxy vitamin d":   "vitamin_d",
	"iron":                   "iron",
	"iron, serum":            "iron",
}

// NormalizeMarkerName maps a raw marker name to a canonical measurement type.
// Returns the original name lowercased if no alias is found.
func NormalizeMarkerName(name string) string {
	lower := toLowerTrimmed(name)
	if canonical, ok := labMarkerAliases[lower]; ok {
		return canonical
	}
	return lower
}

// DetermineFlag computes the out-of-range flag for a measurement value.
// Returns nil if both reference bounds are nil.
func DetermineFlag(value float64, refLow, refHigh *float64) *MeasurementFlag {
	if refLow == nil && refHigh == nil {
		return nil
	}
	flag := FlagNormal
	if refLow != nil && value < *refLow {
		flag = FlagLow
	}
	if refHigh != nil && value > *refHigh {
		flag = FlagHigh
	}
	return &flag
}

// toLowerTrimmed returns a trimmed, lowercased string.
func toLowerTrimmed(s string) string {
	result := make([]byte, 0, len(s))
	// Trim leading/trailing whitespace
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	for i := start; i < end; i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result = append(result, c+32)
		} else {
			result = append(result, c)
		}
	}
	return string(result)
}

// ParseReferenceRange parses a reference range string into low and high bounds.
// Supports formats: "70-100", "<200", ">40", "< 5.7", ">= 30".
func ParseReferenceRange(rangeStr string) (low *float64, high *float64) {
	s := toLowerTrimmed(rangeStr)
	if s == "" {
		return nil, nil
	}

	// Handle "<" prefix (e.g., "<200", "< 100", "<=100")
	if s[0] == '<' {
		s = s[1:]
		if len(s) > 0 && s[0] == '=' {
			s = s[1:]
		}
		s = toLowerTrimmed(s)
		if v, ok := parseFloat(s); ok {
			return nil, &v
		}
		return nil, nil
	}

	// Handle ">" prefix (e.g., ">40", "> 50", ">=30")
	if s[0] == '>' {
		s = s[1:]
		if len(s) > 0 && s[0] == '=' {
			s = s[1:]
		}
		s = toLowerTrimmed(s)
		if v, ok := parseFloat(s); ok {
			return &v, nil
		}
		return nil, nil
	}

	// Handle range format "low-high" (e.g., "70-100", "0.4-4.0")
	dashIdx := findRangeDash(s)
	if dashIdx > 0 {
		lowStr := toLowerTrimmed(s[:dashIdx])
		highStr := toLowerTrimmed(s[dashIdx+1:])
		var lowVal, highVal *float64
		if v, ok := parseFloat(lowStr); ok {
			lowVal = &v
		}
		if v, ok := parseFloat(highStr); ok {
			highVal = &v
		}
		return lowVal, highVal
	}

	return nil, nil
}

// findRangeDash finds the index of the dash separating range values.
// Skips a leading minus sign (negative number).
func findRangeDash(s string) int {
	start := 0
	if len(s) > 0 && s[0] == '-' {
		start = 1
	}
	for i := start; i < len(s); i++ {
		if s[i] == '-' {
			return i
		}
	}
	return -1
}

// parseFloat parses a string to float64 without importing strconv in domain.
func parseFloat(s string) (float64, bool) {
	if s == "" {
		return 0, false
	}

	negative := false
	i := 0
	if s[0] == '-' {
		negative = true
		i = 1
	} else if s[0] == '+' {
		i = 1
	}

	var intPart float64
	hasDigit := false
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		intPart = intPart*10 + float64(s[i]-'0')
		hasDigit = true
		i++
	}

	var fracPart float64
	if i < len(s) && s[i] == '.' {
		i++
		divisor := 10.0
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			fracPart += float64(s[i]-'0') / divisor
			divisor *= 10
			hasDigit = true
			i++
		}
	}

	if !hasDigit || i != len(s) {
		return 0, false
	}

	result := intPart + fracPart
	if negative {
		result = -result
	}
	return result, true
}
