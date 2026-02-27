package entities

import (
	"regexp"
	"strconv"
	"strings"
)

// totalInBodyMetrics is the total number of InBody metrics we attempt to extract.
const totalInBodyMetrics = 6

// metricPattern defines a regex pattern for extracting a specific InBody metric.
type metricPattern struct {
	mtype   MeasurementType
	pattern *regexp.Regexp
	// unitIdx and valIdx specify which capture groups hold the value and unit.
	valIdx  int
	unitIdx int
}

// inBodyPatterns are regex patterns for each InBody metric.
// Each pattern captures a numeric value and its unit from InBody-formatted text.
var inBodyPatterns = []metricPattern{
	{
		mtype:   MeasurementTypeWeight,
		pattern: regexp.MustCompile(`(?i)(?:^|\b)Weight\s*[:=\t]?\s*([\d]+\.?[\d]*)\s*(kg|lbs?|lb)\b`),
		valIdx:  1, unitIdx: 2,
	},
	{
		mtype:   MeasurementTypeBodyFatPct,
		pattern: regexp.MustCompile(`(?i)(?:Body\s*Fat(?:\s*(?:Percentage|Percent|%|Pct))?|PBF)\s*[:=\t]?\s*([\d]+\.?[\d]*)\s*(%)?`),
		valIdx:  1, unitIdx: 0,
	},
	{
		mtype:   MeasurementTypeSkeletalMuscleMass,
		pattern: regexp.MustCompile(`(?i)(?:Skeletal\s*Muscle\s*Mass|SMM)\s*[:=\t]?\s*([\d]+\.?[\d]*)\s*(kg|lbs?|lb)\b`),
		valIdx:  1, unitIdx: 2,
	},
	{
		mtype:   MeasurementTypeBMR,
		pattern: regexp.MustCompile(`(?i)(?:Basal\s*Metabolic\s*Rate|BMR)\s*[:=\t]?\s*([\d]+\.?[\d]*)\s*(kcal|cal)?\b`),
		valIdx:  1, unitIdx: 2,
	},
	{
		mtype:   MeasurementTypeTotalBodyWater,
		pattern: regexp.MustCompile(`(?i)(?:Total\s*Body\s*Water|TBW)\s*[:=\t]?\s*([\d]+\.?[\d]*)\s*(L|l|liters?)?\b`),
		valIdx:  1, unitIdx: 2,
	},
	{
		mtype:   MeasurementTypeLeanBodyMass,
		pattern: regexp.MustCompile(`(?i)(?:Lean\s*Body\s*Mass|LBM|Fat[\s-]?Free\s*Mass)\s*[:=\t]?\s*([\d]+\.?[\d]*)\s*(kg|lbs?|lb)\b`),
		valIdx:  1, unitIdx: 2,
	},
}

// ParseInBodyText extracts body composition metrics from InBody-formatted text.
// Returns an error if the text is empty or no metrics could be extracted.
// Sets Partial=true if fewer than all expected metrics were found.
func ParseInBodyText(text string) (*InBodyResult, error) {
	if strings.TrimSpace(text) == "" {
		return nil, NewValidationError("empty text: cannot extract InBody metrics")
	}

	seen := map[MeasurementType]bool{}
	var metrics []ExtractedMetric

	for _, mp := range inBodyPatterns {
		if seen[mp.mtype] {
			continue
		}
		match := mp.pattern.FindStringSubmatch(text)
		if match == nil {
			continue
		}

		val, err := strconv.ParseFloat(match[mp.valIdx], 64)
		if err != nil {
			continue
		}

		unit := resolveUnit(mp, match)
		metrics = append(metrics, ExtractedMetric{
			Type:  mp.mtype,
			Value: val,
			Unit:  unit,
		})
		seen[mp.mtype] = true
	}

	if len(metrics) == 0 {
		return nil, NewValidationError("no InBody metrics found in text")
	}

	return &InBodyResult{
		Metrics: metrics,
		Partial: len(metrics) < totalInBodyMetrics,
	}, nil
}

// resolveUnit extracts and normalizes the unit from a regex match.
func resolveUnit(mp metricPattern, match []string) string {
	// Body fat % always uses % as unit
	if mp.mtype == MeasurementTypeBodyFatPct {
		return "%"
	}
	// BMR defaults to kcal
	if mp.unitIdx > 0 && mp.unitIdx < len(match) && match[mp.unitIdx] != "" {
		return normalizeUnit(match[mp.unitIdx])
	}
	if mp.mtype == MeasurementTypeBMR {
		return "kcal"
	}
	if mp.mtype == MeasurementTypeTotalBodyWater {
		return "L"
	}
	return ""
}

// normalizeUnit standardizes unit strings to canonical forms.
func normalizeUnit(unit string) string {
	lower := strings.ToLower(strings.TrimSpace(unit))
	switch lower {
	case "lb", "lbs":
		return "lbs"
	case "kg":
		return "kg"
	case "kcal", "cal":
		return "kcal"
	case "l", "liters", "liter":
		return "L"
	case "%":
		return "%"
	default:
		return unit
	}
}
