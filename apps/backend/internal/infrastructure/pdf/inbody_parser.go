package pdf

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// InBodyParser implements repository.InBodyTextParser using regex-based extraction.
type InBodyParser struct{}

// NewInBodyParser creates a new InBodyParser.
func NewInBodyParser() *InBodyParser {
	return &InBodyParser{}
}

// Parse extracts InBody body composition metrics from raw text.
func (p *InBodyParser) Parse(text, clientID, recordedBy, artifactID string, measuredAt time.Time) *entities.ExtractionResult {
	return ParseInBodyText(text, clientID, recordedBy, artifactID, measuredAt)
}

// Compile-time interface check.
var _ repository.InBodyTextParser = &InBodyParser{}

// inbodyField defines how to extract a single field from InBody PDF text.
type inbodyField struct {
	MeasurementType entities.MeasurementType
	Patterns        []*regexp.Regexp
	Unit            string
}

// inbodyFields lists the fields we extract from InBody PDFs with their regex patterns.
// Multiple patterns support different InBody models (270, 570, 770, 970).
var inbodyFields = []inbodyField{
	{
		MeasurementType: entities.MeasurementTypeWeight,
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:body\s*weight|weight)\s*[:\-]?\s*([\d]+\.?\d*)\s*(kg|lbs?|lb)`),
		},
		Unit: "kg",
	},
	{
		MeasurementType: entities.MeasurementTypeBodyFatPct,
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:percent\s*body\s*fat|body\s*fat\s*(?:percentage|pct|%))\s*[:\-]?\s*([\d]+\.?\d*)\s*%?`),
			regexp.MustCompile(`(?i)PBF\s*[:\-]?\s*([\d]+\.?\d*)\s*%?`),
		},
		Unit: "%",
	},
	{
		MeasurementType: entities.MeasurementTypeSkeletalMuscleMass,
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:skeletal\s*muscle\s*mass|SMM)\s*[:\-]?\s*([\d]+\.?\d*)\s*(kg|lbs?|lb)`),
		},
		Unit: "kg",
	},
	{
		MeasurementType: entities.MeasurementTypeBMR,
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:basal\s*metabolic\s*rate|BMR)\s*[:\-]?\s*([\d]+\.?\d*)\s*(?:kcal)?`),
		},
		Unit: "kcal",
	},
	{
		MeasurementType: entities.MeasurementTypeTotalBodyWater,
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:total\s*body\s*water|TBW)\s*[:\-]?\s*([\d]+\.?\d*)\s*(?:L|liter)?`),
		},
		Unit: "L",
	},
	{
		MeasurementType: entities.MeasurementTypeLeanBodyMass,
		Patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(?:lean\s*body\s*mass|LBM|fat\s*free\s*mass|FFM)\s*[:\-]?\s*([\d]+\.?\d*)\s*(kg|lbs?|lb)`),
		},
		Unit: "kg",
	},
}

// ParseInBodyText extracts InBody metrics from raw text content.
// It returns an ExtractionResult containing all found measurements.
func ParseInBodyText(text, clientID, recordedBy, artifactID string, measuredAt time.Time) *entities.ExtractionResult {
	result := &entities.ExtractionResult{}

	for _, field := range inbodyFields {
		value, unit, found := matchField(text, field)
		if !found {
			result.Errors = append(result.Errors, "field not found: "+string(field.MeasurementType))
			continue
		}

		result.Measurements = append(result.Measurements, &entities.Measurement{
			ClientID:         clientID,
			RecordedBy:       recordedBy,
			MeasurementType:  field.MeasurementType,
			Value:            value,
			Unit:             unit,
			MeasuredAt:       measuredAt,
			ArtifactID:       artifactID,
			ExtractionStatus: entities.ExtractionStatusComplete,
		})
	}

	totalFields := len(inbodyFields)
	extracted := len(result.Measurements)
	if extracted > 0 && extracted < totalFields {
		result.IsPartial = true
		for _, m := range result.Measurements {
			m.ExtractionStatus = entities.ExtractionStatusPartial
		}
	}

	return result
}

// matchField tries all patterns for a field and returns the first match.
func matchField(text string, field inbodyField) (float64, string, bool) {
	for _, pattern := range field.Patterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) < 2 {
			continue
		}
		value, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			continue
		}
		unit := field.Unit
		if len(matches) >= 3 && matches[2] != "" {
			unit = normalizeUnit(matches[2])
		}
		return value, unit, true
	}
	return 0, "", false
}

// normalizeUnit standardizes unit strings.
func normalizeUnit(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	switch s {
	case "lb", "lbs":
		return "lbs"
	case "kg":
		return "kg"
	case "l", "liter", "liters":
		return "L"
	case "kcal":
		return "kcal"
	default:
		return s
	}
}
