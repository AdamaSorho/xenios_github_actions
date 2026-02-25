package repository

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/xenios/backend/internal/domain/entities"
)

// InBodyTextExtractor extracts InBody metrics from PDF text content.
// It uses regex-based heuristics that work across InBody 270, 570, 770, and 970 models.
type InBodyTextExtractor struct {
	textExtractor PDFTextExtractorFunc
}

// PDFTextExtractorFunc converts raw PDF bytes into plain text.
// This is injected to allow testing with plain text without actual PDF parsing.
type PDFTextExtractorFunc func(pdfBytes []byte) (string, error)

// NewInBodyTextExtractor creates a new extractor with the given PDF-to-text function.
func NewInBodyTextExtractor(textExtractor PDFTextExtractorFunc) *InBodyTextExtractor {
	return &InBodyTextExtractor{textExtractor: textExtractor}
}

// ExtractInBody extracts InBody metrics from PDF content.
func (e *InBodyTextExtractor) ExtractInBody(_ context.Context, pdfContent []byte) (*entities.InBodyResult, error) {
	if len(pdfContent) == 0 {
		return nil, fmt.Errorf("empty PDF content")
	}

	text, err := e.textExtractor(pdfContent)
	if err != nil {
		return nil, fmt.Errorf("extract text from PDF: %w", err)
	}

	if text == "" {
		return nil, fmt.Errorf("no text content extracted from PDF")
	}

	result := &entities.InBodyResult{}

	result.Weight, result.WeightUnit = extractWeight(text)
	result.BodyFatPct = extractBodyFatPct(text)
	result.SkeletalMuscleMass, result.SMMUnit = extractSkeletalMuscleMass(text)
	result.BMR = extractBMR(text)
	result.TotalBodyWater = extractTotalBodyWater(text)
	result.LeanBodyMass, result.LBMUnit = extractLeanBodyMass(text)

	return result, nil
}

// Patterns for InBody metric extraction.
// InBody reports use consistent labeling across models (270, 570, 770, 970).

var weightPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:body\s*)?weight\s*[:\s]+(\d+\.?\d*)\s*(kg|lbs?|pounds?)?`),
	regexp.MustCompile(`(?i)weight\s*\(?(kg|lbs?)\)?\s*[:\s]+(\d+\.?\d*)`),
}

var bodyFatPctPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:percent\s*)?body\s*fat\s*(?:percentage|percent|pct)?\s*[:\s]+(\d+\.?\d*)\s*%?`),
	regexp.MustCompile(`(?i)PBF\s*[:\s]+(\d+\.?\d*)\s*%?`),
	regexp.MustCompile(`(?i)body\s*fat\s*%\s*[:\s]+(\d+\.?\d*)`),
	regexp.MustCompile(`(?i)BFM\s*%\s*[:\s]+(\d+\.?\d*)`),
}

var smmPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)skeletal\s*muscle\s*mass\s*[:\s]+(\d+\.?\d*)\s*(kg|lbs?)?`),
	regexp.MustCompile(`(?i)SMM\s*[:\s]+(\d+\.?\d*)\s*(kg|lbs?)?`),
}

var bmrPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:basal\s*)?(?:metabolic\s*)?rate\s*[:\s]+(\d+\.?\d*)\s*(?:kcal)?`),
	regexp.MustCompile(`(?i)BMR\s*[:\s]+(\d+\.?\d*)\s*(?:kcal)?`),
}

var tbwPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)total\s*body\s*water\s*[:\s]+(\d+\.?\d*)\s*(?:L|liters?)?`),
	regexp.MustCompile(`(?i)TBW\s*[:\s]+(\d+\.?\d*)\s*(?:L|liters?)?`),
}

var lbmPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:lean|fat[\s-]*free)\s*body\s*mass\s*[:\s]+(\d+\.?\d*)\s*(kg|lbs?)?`),
	regexp.MustCompile(`(?i)LBM\s*[:\s]+(\d+\.?\d*)\s*(kg|lbs?)?`),
	regexp.MustCompile(`(?i)FFM\s*[:\s]+(\d+\.?\d*)\s*(kg|lbs?)?`),
}

func extractWeight(text string) (*float64, string) {
	for _, p := range weightPatterns {
		matches := p.FindStringSubmatch(text)
		if matches == nil {
			continue
		}
		// Pattern 1: value in group 1, unit in group 2
		// Pattern 2: unit in group 1, value in group 2
		var valStr, unitStr string
		if len(matches) >= 3 {
			// Check if first group is a number
			if _, err := strconv.ParseFloat(matches[1], 64); err == nil {
				valStr = matches[1]
				unitStr = matches[2]
			} else {
				unitStr = matches[1]
				valStr = matches[2]
			}
		} else if len(matches) >= 2 {
			valStr = matches[1]
		}

		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			continue
		}
		unit := normalizeUnit(unitStr, "kg")
		return &val, unit
	}
	return nil, ""
}

func extractBodyFatPct(text string) *float64 {
	for _, p := range bodyFatPctPatterns {
		matches := p.FindStringSubmatch(text)
		if matches != nil && len(matches) >= 2 {
			val, err := strconv.ParseFloat(matches[1], 64)
			if err == nil && val >= 0 && val <= 100 {
				return &val
			}
		}
	}
	return nil
}

func extractSkeletalMuscleMass(text string) (*float64, string) {
	for _, p := range smmPatterns {
		matches := p.FindStringSubmatch(text)
		if matches != nil && len(matches) >= 2 {
			val, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				unit := "kg"
				if len(matches) >= 3 {
					unit = normalizeUnit(matches[2], "kg")
				}
				return &val, unit
			}
		}
	}
	return nil, ""
}

func extractBMR(text string) *float64 {
	for _, p := range bmrPatterns {
		matches := p.FindStringSubmatch(text)
		if matches != nil && len(matches) >= 2 {
			val, err := strconv.ParseFloat(matches[1], 64)
			if err == nil && val > 0 {
				return &val
			}
		}
	}
	return nil
}

func extractTotalBodyWater(text string) *float64 {
	for _, p := range tbwPatterns {
		matches := p.FindStringSubmatch(text)
		if matches != nil && len(matches) >= 2 {
			val, err := strconv.ParseFloat(matches[1], 64)
			if err == nil && val > 0 {
				return &val
			}
		}
	}
	return nil
}

func extractLeanBodyMass(text string) (*float64, string) {
	for _, p := range lbmPatterns {
		matches := p.FindStringSubmatch(text)
		if matches != nil && len(matches) >= 2 {
			val, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				unit := "kg"
				if len(matches) >= 3 {
					unit = normalizeUnit(matches[2], "kg")
				}
				return &val, unit
			}
		}
	}
	return nil, ""
}

func normalizeUnit(raw, defaultUnit string) string {
	lower := strings.TrimSpace(strings.ToLower(raw))
	switch lower {
	case "kg":
		return "kg"
	case "lb", "lbs", "pound", "pounds":
		return "lbs"
	case "l", "liter", "liters":
		return "L"
	case "kcal":
		return "kcal"
	case "%":
		return "%"
	default:
		return defaultUnit
	}
}
