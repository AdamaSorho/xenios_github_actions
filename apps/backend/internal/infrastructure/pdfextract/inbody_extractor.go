package pdfextract

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/xenios/backend/internal/domain/entities"
)

// InBodyExtractor extracts body composition measurements from InBody PDF text.
// It uses regex-based heuristics to parse the structured InBody report layout.
type InBodyExtractor struct{}

// NewInBodyExtractor creates a new InBodyExtractor.
func NewInBodyExtractor() *InBodyExtractor {
	return &InBodyExtractor{}
}

// fieldPattern defines a regex pattern and metadata for a single InBody field.
type fieldPattern struct {
	measurementType entities.MeasurementType
	patterns        []*regexp.Regexp
	unit            string
}

// inBodyPatterns contains regex patterns for extracting InBody fields.
// Multiple patterns per field handle different InBody models (270, 570, 770, 970).
var inBodyPatterns = []fieldPattern{
	{
		measurementType: entities.MeasurementTypeWeight,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:body\s*)?weight\s*[:\-]?\s*(\d+\.?\d*)\s*(kg|lbs?|pounds?)`),
			regexp.MustCompile(`(?i)weight\s*\(?(kg|lbs?)\)?\s*[:\-]?\s*(\d+\.?\d*)`),
		},
		unit: "", // detected from match
	},
	{
		measurementType: entities.MeasurementTypeBodyFatPct,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:percent\s*)?body\s*fat\s*(?:percentage|%|pct)?\s*[:\-]?\s*(\d+\.?\d*)\s*%?`),
			regexp.MustCompile(`(?i)PBF\s*[:\-]?\s*(\d+\.?\d*)\s*%?`),
			regexp.MustCompile(`(?i)body\s*fat\s*%?\s*[:\-]?\s*(\d+\.?\d*)`),
		},
		unit: "%",
	},
	{
		measurementType: entities.MeasurementTypeSkeletalMuscleMass,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)skeletal\s*muscle\s*mass\s*[:\-]?\s*(\d+\.?\d*)\s*(kg|lbs?)?`),
			regexp.MustCompile(`(?i)SMM\s*[:\-]?\s*(\d+\.?\d*)\s*(kg|lbs?)?`),
		},
		unit: "", // detected from match
	},
	{
		measurementType: entities.MeasurementTypeBMR,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:basal\s*)?(?:metabolic\s*rate|BMR)\s*[:\-]?\s*(\d+\.?\d*)\s*(?:kcal|cal)?`),
		},
		unit: "kcal",
	},
	{
		measurementType: entities.MeasurementTypeTotalBodyWater,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)total\s*body\s*water\s*[:\-]?\s*(\d+\.?\d*)\s*(?:L|liters?)?`),
			regexp.MustCompile(`(?i)TBW\s*[:\-]?\s*(\d+\.?\d*)\s*(?:L|liters?)?`),
		},
		unit: "L",
	},
	{
		measurementType: entities.MeasurementTypeLeanBodyMass,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:lean\s*body\s*mass|fat\s*free\s*mass)\s*[:\-]?\s*(\d+\.?\d*)\s*(kg|lbs?)?`),
			regexp.MustCompile(`(?i)LBM\s*[:\-]?\s*(\d+\.?\d*)\s*(kg|lbs?)?`),
			regexp.MustCompile(`(?i)FFM\s*[:\-]?\s*(\d+\.?\d*)\s*(kg|lbs?)?`),
		},
		unit: "", // detected from match
	},
}

// ExtractInBody parses InBody PDF text content and returns extracted measurements.
func (e *InBodyExtractor) ExtractInBody(_ context.Context, pdfData []byte) (*entities.ExtractionResult, error) {
	text := string(pdfData)
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("empty PDF content")
	}

	result := &entities.ExtractionResult{
		Measurements: make([]*entities.Measurement, 0),
	}

	for _, fp := range inBodyPatterns {
		value, unit, err := extractField(text, fp)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", fp.measurementType, err))
			continue
		}

		m := &entities.Measurement{
			MeasurementType: fp.measurementType,
			Value:           value,
			Unit:            unit,
		}
		result.Measurements = append(result.Measurements, m)
	}

	if len(result.Measurements) == 0 {
		return nil, fmt.Errorf("no InBody metrics could be extracted from PDF")
	}

	totalFields := len(inBodyPatterns)
	if len(result.Measurements) < totalFields {
		result.Partial = true
	}

	return result, nil
}

// extractField attempts to extract a single field value from text using the field's patterns.
func extractField(text string, fp fieldPattern) (float64, string, error) {
	for _, pat := range fp.patterns {
		matches := pat.FindStringSubmatch(text)
		if matches == nil {
			continue
		}

		value, unit, err := parseMatch(matches, fp)
		if err != nil {
			continue
		}
		return value, unit, nil
	}

	return 0, "", fmt.Errorf("field not found")
}

// parseMatch extracts the numeric value and unit from regex match groups.
func parseMatch(matches []string, fp fieldPattern) (float64, string, error) {
	// Try each capture group for a valid number
	for i := 1; i < len(matches); i++ {
		val, err := strconv.ParseFloat(matches[i], 64)
		if err == nil {
			unit := fp.unit
			if unit == "" {
				unit = detectUnit(matches, fp.measurementType)
			}
			return val, unit, nil
		}
	}
	return 0, "", fmt.Errorf("no numeric value found")
}

// detectUnit inspects match groups for a unit string.
func detectUnit(matches []string, mt entities.MeasurementType) string {
	for _, m := range matches[1:] {
		lower := strings.ToLower(strings.TrimSpace(m))
		switch lower {
		case "kg":
			return "kg"
		case "lb", "lbs", "pounds", "pound":
			return "lbs"
		case "l", "liters", "liter":
			return "L"
		}
	}

	// Default units based on measurement type
	switch mt {
	case entities.MeasurementTypeWeight, entities.MeasurementTypeSkeletalMuscleMass, entities.MeasurementTypeLeanBodyMass:
		return "kg"
	case entities.MeasurementTypeTotalBodyWater:
		return "L"
	default:
		return ""
	}
}
