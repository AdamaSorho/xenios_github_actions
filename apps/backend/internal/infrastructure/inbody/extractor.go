package inbody

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// TextExtractor implements PDFExtractor by parsing raw text from InBody PDF reports.
// InBody scans follow a consistent layout, so regex-based extraction works for MVP.
type TextExtractor struct {
	textExtractFn func(pdfData []byte) (string, error)
}

// NewTextExtractor creates a TextExtractor with the given text extraction function.
// The textExtractFn converts raw PDF bytes into a plain text string.
func NewTextExtractor(textExtractFn func(pdfData []byte) (string, error)) *TextExtractor {
	return &TextExtractor{textExtractFn: textExtractFn}
}

// ExtractInBody parses PDF content and returns structured InBody results.
func (e *TextExtractor) ExtractInBody(_ context.Context, pdfData []byte) (*entities.InBodyResult, error) {
	if len(pdfData) == 0 {
		return nil, fmt.Errorf("empty PDF data")
	}

	text, err := e.textExtractFn(pdfData)
	if err != nil {
		return nil, fmt.Errorf("extract text from PDF: %w", err)
	}

	if text == "" {
		return nil, fmt.Errorf("no text extracted from PDF")
	}

	result := &entities.InBodyResult{}

	result.Weight, result.WeightUnit = extractWeight(text)
	result.BodyFatPct = extractBodyFatPct(text)
	result.SkeletalMuscleMass, result.SkeletalMuscleUnit = extractSkeletalMuscleMass(text)
	result.BMR = extractBMR(text)
	result.TotalBodyWater = extractTotalBodyWater(text)
	result.LeanBodyMass, result.LeanBodyMassUnit = extractLeanBodyMass(text)
	result.MeasuredAt = extractDate(text)

	if result.FieldCount() == 0 {
		return nil, fmt.Errorf("no InBody metrics found in PDF text")
	}

	result.IsPartial = result.FieldCount() < entities.TotalFields

	return result, nil
}

// float64Ptr returns a pointer to a float64 value.
func float64Ptr(v float64) *float64 {
	return &v
}

// extractWeight looks for weight patterns in InBody text.
var weightPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:body\s*weight|weight)\s*[:=]?\s*(\d+\.?\d*)\s*(kg|lbs?|lb)`),
	regexp.MustCompile(`(?i)weight\s*\(?(kg|lbs?)\)?\s*[:=]?\s*(\d+\.?\d*)`),
}

func extractWeight(text string) (*float64, string) {
	for _, p := range weightPatterns {
		m := p.FindStringSubmatch(text)
		if m == nil {
			continue
		}
		// Pattern 1: value then unit
		if len(m) >= 3 {
			val, err := strconv.ParseFloat(m[1], 64)
			if err == nil && val > 0 {
				return float64Ptr(val), normalizeWeightUnit(m[2])
			}
			// Pattern 2: unit then value
			val, err = strconv.ParseFloat(m[2], 64)
			if err == nil && val > 0 {
				return float64Ptr(val), normalizeWeightUnit(m[1])
			}
		}
	}
	return nil, ""
}

// extractBodyFatPct looks for body fat percentage patterns.
var bodyFatPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:body\s*fat\s*(?:percentage|pct|%))\s*[:=]?\s*(\d+\.?\d*)`),
	regexp.MustCompile(`(?i)(?:percent\s*body\s*fat|PBF)\s*[:=]?\s*(\d+\.?\d*)`),
	regexp.MustCompile(`(?i)body\s*fat\s*[:=]?\s*(\d+\.?\d*)\s*%`),
}

func extractBodyFatPct(text string) *float64 {
	for _, p := range bodyFatPatterns {
		m := p.FindStringSubmatch(text)
		if m == nil || len(m) < 2 {
			continue
		}
		val, err := strconv.ParseFloat(m[1], 64)
		if err == nil && val >= 0 && val <= 100 {
			return float64Ptr(val)
		}
	}
	return nil
}

// extractSkeletalMuscleMass looks for skeletal muscle mass patterns.
var smmPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:skeletal\s*muscle\s*mass|SMM)\s*[:=]?\s*(\d+\.?\d*)\s*(kg|lbs?|lb)`),
	regexp.MustCompile(`(?i)(?:skeletal\s*muscle\s*mass|SMM)\s*\(?(kg|lbs?)\)?\s*[:=]?\s*(\d+\.?\d*)`),
}

func extractSkeletalMuscleMass(text string) (*float64, string) {
	for _, p := range smmPatterns {
		m := p.FindStringSubmatch(text)
		if m == nil || len(m) < 3 {
			continue
		}
		val, err := strconv.ParseFloat(m[1], 64)
		if err == nil && val > 0 {
			return float64Ptr(val), normalizeWeightUnit(m[2])
		}
		val, err = strconv.ParseFloat(m[2], 64)
		if err == nil && val > 0 {
			return float64Ptr(val), normalizeWeightUnit(m[1])
		}
	}
	return nil, ""
}

// extractBMR looks for basal metabolic rate patterns.
var bmrPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:basal\s*metabolic\s*rate|BMR)\s*[:=]?\s*(\d+\.?\d*)\s*(?:kcal)?`),
	regexp.MustCompile(`(?i)BMR\s*\(?kcal\)?\s*[:=]?\s*(\d+\.?\d*)`),
}

func extractBMR(text string) *float64 {
	for _, p := range bmrPatterns {
		m := p.FindStringSubmatch(text)
		if m == nil || len(m) < 2 {
			continue
		}
		val, err := strconv.ParseFloat(m[1], 64)
		if err == nil && val > 0 {
			return float64Ptr(val)
		}
	}
	return nil
}

// extractTotalBodyWater looks for total body water patterns.
var tbwPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:total\s*body\s*water|TBW)\s*[:=]?\s*(\d+\.?\d*)\s*(?:L|l|liters?)?`),
	regexp.MustCompile(`(?i)TBW\s*\(?L\)?\s*[:=]?\s*(\d+\.?\d*)`),
}

func extractTotalBodyWater(text string) *float64 {
	for _, p := range tbwPatterns {
		m := p.FindStringSubmatch(text)
		if m == nil || len(m) < 2 {
			continue
		}
		val, err := strconv.ParseFloat(m[1], 64)
		if err == nil && val > 0 {
			return float64Ptr(val)
		}
	}
	return nil
}

// extractLeanBodyMass looks for lean body mass patterns.
var lbmPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:lean\s*body\s*mass|LBM|fat[- ]?free\s*mass|FFM)\s*[:=]?\s*(\d+\.?\d*)\s*(kg|lbs?|lb)`),
	regexp.MustCompile(`(?i)(?:lean\s*body\s*mass|LBM|fat[- ]?free\s*mass|FFM)\s*\(?(kg|lbs?)\)?\s*[:=]?\s*(\d+\.?\d*)`),
}

func extractLeanBodyMass(text string) (*float64, string) {
	for _, p := range lbmPatterns {
		m := p.FindStringSubmatch(text)
		if m == nil || len(m) < 3 {
			continue
		}
		val, err := strconv.ParseFloat(m[1], 64)
		if err == nil && val > 0 {
			return float64Ptr(val), normalizeWeightUnit(m[2])
		}
		val, err = strconv.ParseFloat(m[2], 64)
		if err == nil && val > 0 {
			return float64Ptr(val), normalizeWeightUnit(m[1])
		}
	}
	return nil, ""
}

// extractDate tries to find a date in the InBody text.
var datePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:test\s*date|date)\s*[:=]?\s*(\d{1,2})[/\-](\d{1,2})[/\-](\d{4})`),
	regexp.MustCompile(`(?i)(\d{4})[/\-](\d{1,2})[/\-](\d{1,2})`),
}

func extractDate(text string) time.Time {
	for _, p := range datePatterns {
		m := p.FindStringSubmatch(text)
		if m == nil || len(m) < 4 {
			continue
		}
		// Try MM/DD/YYYY first
		month, _ := strconv.Atoi(m[1])
		day, _ := strconv.Atoi(m[2])
		year, _ := strconv.Atoi(m[3])

		// Detect YYYY-MM-DD format
		if m[1] != "" && len(m[1]) == 4 {
			year, _ = strconv.Atoi(m[1])
			month, _ = strconv.Atoi(m[2])
			day, _ = strconv.Atoi(m[3])
		}

		if year > 0 && month >= 1 && month <= 12 && day >= 1 && day <= 31 {
			return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		}
	}
	return time.Now().UTC().Truncate(24 * time.Hour)
}

// normalizeWeightUnit normalizes weight unit strings.
func normalizeWeightUnit(unit string) string {
	u := strings.ToLower(strings.TrimSpace(unit))
	switch u {
	case "lb", "lbs":
		return "lbs"
	case "kg":
		return "kg"
	default:
		return u
	}
}
