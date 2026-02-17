package pdfextract

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestExtractInBody_FullReport_ExtractsAllFields(t *testing.T) {
	text := `
InBody Result Sheet
ID: 12345
Date: 2024-01-15

Body Composition Analysis
Weight: 85.4 kg
Skeletal Muscle Mass: 38.2 kg
Body Fat Percentage: 22.3%
Total Body Water: 42.1 L
Lean Body Mass: 66.5 kg

Basal Metabolic Rate: 1847 kcal
`
	extractor := NewInBodyExtractor()
	result, err := extractor.ExtractInBody(context.Background(), []byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Partial {
		t.Errorf("expected full extraction, got partial; errors: %v", result.Errors)
	}

	expected := map[entities.MeasurementType]struct {
		value float64
		unit  string
	}{
		entities.MeasurementTypeWeight:             {85.4, "kg"},
		entities.MeasurementTypeSkeletalMuscleMass: {38.2, "kg"},
		entities.MeasurementTypeBodyFatPct:         {22.3, "%"},
		entities.MeasurementTypeTotalBodyWater:     {42.1, "L"},
		entities.MeasurementTypeLeanBodyMass:       {66.5, "kg"},
		entities.MeasurementTypeBMR:                {1847, "kcal"},
	}

	if len(result.Measurements) != len(expected) {
		t.Fatalf("expected %d measurements, got %d", len(expected), len(result.Measurements))
	}

	for _, m := range result.Measurements {
		exp, ok := expected[m.MeasurementType]
		if !ok {
			t.Errorf("unexpected measurement type: %s", m.MeasurementType)
			continue
		}
		if m.Value != exp.value {
			t.Errorf("%s: expected value %.1f, got %.1f", m.MeasurementType, exp.value, m.Value)
		}
		if m.Unit != exp.unit {
			t.Errorf("%s: expected unit %q, got %q", m.MeasurementType, exp.unit, m.Unit)
		}
	}
}

func TestExtractInBody_InBody570Format_ExtractsFields(t *testing.T) {
	text := `
InBody 570 Body Composition Analysis

Weight (kg): 92.1
SMM: 42.5 kg
PBF: 18.7 %
TBW: 48.3 L
LBM: 74.8 kg
BMR: 2015 kcal
`
	extractor := NewInBodyExtractor()
	result, err := extractor.ExtractInBody(context.Background(), []byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Partial {
		t.Errorf("expected full extraction, got partial; errors: %v", result.Errors)
	}

	if len(result.Measurements) != 6 {
		t.Errorf("expected 6 measurements, got %d", len(result.Measurements))
	}

	// Verify specific values
	for _, m := range result.Measurements {
		switch m.MeasurementType {
		case entities.MeasurementTypeWeight:
			if m.Value != 92.1 {
				t.Errorf("weight: expected 92.1, got %.1f", m.Value)
			}
		case entities.MeasurementTypeSkeletalMuscleMass:
			if m.Value != 42.5 {
				t.Errorf("SMM: expected 42.5, got %.1f", m.Value)
			}
		case entities.MeasurementTypeBodyFatPct:
			if m.Value != 18.7 {
				t.Errorf("PBF: expected 18.7, got %.1f", m.Value)
			}
		}
	}
}

func TestExtractInBody_LbsUnits_DetectsCorrectUnit(t *testing.T) {
	text := `
InBody Report
Weight: 185.4 lbs
Skeletal Muscle Mass: 82.3 lbs
Body Fat %: 22.1
Total Body Water: 42.1 L
Lean Body Mass: 144.1 lbs
BMR: 1900 kcal
`
	extractor := NewInBodyExtractor()
	result, err := extractor.ExtractInBody(context.Background(), []byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, m := range result.Measurements {
		switch m.MeasurementType {
		case entities.MeasurementTypeWeight:
			if m.Unit != "lbs" {
				t.Errorf("weight: expected unit 'lbs', got %q", m.Unit)
			}
			if m.Value != 185.4 {
				t.Errorf("weight: expected 185.4, got %.1f", m.Value)
			}
		case entities.MeasurementTypeSkeletalMuscleMass:
			if m.Unit != "lbs" {
				t.Errorf("SMM: expected unit 'lbs', got %q", m.Unit)
			}
		}
	}
}

func TestExtractInBody_PartialExtraction_FlaggedAsPartial(t *testing.T) {
	text := `
InBody Report
Weight: 85.4 kg
Body Fat Percentage: 22.3%
`
	extractor := NewInBodyExtractor()
	result, err := extractor.ExtractInBody(context.Background(), []byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Partial {
		t.Error("expected partial extraction")
	}

	if len(result.Measurements) != 2 {
		t.Errorf("expected 2 measurements, got %d", len(result.Measurements))
	}

	if len(result.Errors) == 0 {
		t.Error("expected errors for missing fields")
	}
}

func TestExtractInBody_EmptyContent_ReturnsError(t *testing.T) {
	extractor := NewInBodyExtractor()

	_, err := extractor.ExtractInBody(context.Background(), []byte(""))
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestExtractInBody_WhitespaceOnly_ReturnsError(t *testing.T) {
	extractor := NewInBodyExtractor()

	_, err := extractor.ExtractInBody(context.Background(), []byte("   \n\t  "))
	if err == nil {
		t.Fatal("expected error for whitespace-only content")
	}
}

func TestExtractInBody_NoMetrics_ReturnsError(t *testing.T) {
	text := `This is not an InBody report. Just random text with no metrics.`
	extractor := NewInBodyExtractor()

	_, err := extractor.ExtractInBody(context.Background(), []byte(text))
	if err == nil {
		t.Fatal("expected error when no metrics found")
	}
}

func TestExtractInBody_InBody270Format_ExtractsFields(t *testing.T) {
	text := `
InBody 270

Body Composition Analysis
Weight - 78.5 kg
Body Fat % - 25.1
Skeletal Muscle Mass - 32.8 kg
Fat Free Mass - 58.8 kg
BMR - 1650 kcal
Total Body Water - 35.2 L
`
	extractor := NewInBodyExtractor()
	result, err := extractor.ExtractInBody(context.Background(), []byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Partial {
		t.Errorf("expected full extraction, got partial; errors: %v", result.Errors)
	}

	if len(result.Measurements) != 6 {
		t.Errorf("expected 6 measurements, got %d", len(result.Measurements))
	}

	// Verify Fat Free Mass maps to LeanBodyMass
	for _, m := range result.Measurements {
		if m.MeasurementType == entities.MeasurementTypeLeanBodyMass {
			if m.Value != 58.8 {
				t.Errorf("LBM (from FFM): expected 58.8, got %.1f", m.Value)
			}
			return
		}
	}
	t.Error("expected lean_body_mass measurement from 'Fat Free Mass'")
}

func TestExtractInBody_InBody770Format_ExtractsFields(t *testing.T) {
	text := `
InBody 770 Result Sheet

Body Composition
Weight: 95.0 kg
Skeletal Muscle Mass: 45.1 kg
Percent Body Fat: 15.2%
Total Body Water: 50.5 L
Lean Body Mass: 80.6 kg
Basal Metabolic Rate: 2100 kcal

Segmental Lean Analysis
Right Arm: 3.8 kg
Left Arm: 3.7 kg
Trunk: 29.1 kg
Right Leg: 10.2 kg
Left Leg: 10.1 kg
`
	extractor := NewInBodyExtractor()
	result, err := extractor.ExtractInBody(context.Background(), []byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Partial {
		t.Errorf("expected full extraction, got partial; errors: %v", result.Errors)
	}

	if len(result.Measurements) != 6 {
		t.Errorf("expected 6 measurements, got %d", len(result.Measurements))
	}
}

func TestExtractField_WeightVariants(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		value float64
		unit  string
	}{
		{"colon separator", "Weight: 80.5 kg", 80.5, "kg"},
		{"dash separator", "Weight - 80.5 kg", 80.5, "kg"},
		{"no separator", "Weight 80.5 kg", 80.5, "kg"},
		{"body weight prefix", "Body Weight: 80.5 kg", 80.5, "kg"},
		{"lbs unit", "Weight: 177.0 lbs", 177.0, "lbs"},
		{"unit in parens", "Weight (kg): 80.5", 80.5, "kg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := inBodyPatterns[0] // weight pattern
			val, unit, err := extractField(tt.text, fp)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val != tt.value {
				t.Errorf("expected value %.1f, got %.1f", tt.value, val)
			}
			if unit != tt.unit {
				t.Errorf("expected unit %q, got %q", tt.unit, unit)
			}
		})
	}
}
