package repository

import (
	"context"
	"fmt"
	"math"
	"testing"
)

// plainTextExtractor is a test helper that returns the raw bytes as text.
func plainTextExtractor(pdfBytes []byte) (string, error) {
	return string(pdfBytes), nil
}

// failingExtractor simulates a PDF text extraction failure.
func failingExtractor(_ []byte) (string, error) {
	return "", fmt.Errorf("failed to parse PDF")
}

// emptyExtractor simulates a PDF with no extractable text.
func emptyExtractor(_ []byte) (string, error) {
	return "", nil
}

func TestInBodyExtractor_EmptyContent_ReturnsError(t *testing.T) {
	ext := NewInBodyTextExtractor(plainTextExtractor)
	_, err := ext.ExtractInBody(context.Background(), []byte{})
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestInBodyExtractor_FailedTextExtraction_ReturnsError(t *testing.T) {
	ext := NewInBodyTextExtractor(failingExtractor)
	_, err := ext.ExtractInBody(context.Background(), []byte("some bytes"))
	if err == nil {
		t.Fatal("expected error for failed text extraction")
	}
}

func TestInBodyExtractor_EmptyText_ReturnsError(t *testing.T) {
	ext := NewInBodyTextExtractor(emptyExtractor)
	_, err := ext.ExtractInBody(context.Background(), []byte("some bytes"))
	if err == nil {
		t.Fatal("expected error for empty text")
	}
}

func TestInBodyExtractor_FullInBody570Report_ExtractsAllFields(t *testing.T) {
	text := `
InBody 570 Body Composition Analysis
Date: 2024-01-15

Body Composition Analysis
Weight: 85.4 kg
Skeletal Muscle Mass: 35.2 kg
Body Fat Percentage: 22.3 %
Total Body Water: 42.1 L
Lean Body Mass: 66.4 kg
BMR: 1847 kcal
`
	ext := NewInBodyTextExtractor(plainTextExtractor)
	result, err := ext.ExtractInBody(context.Background(), []byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFloat(t, "weight", result.Weight, 85.4)
	if result.WeightUnit != "kg" {
		t.Errorf("expected weight unit 'kg', got %q", result.WeightUnit)
	}
	assertFloat(t, "body fat pct", result.BodyFatPct, 22.3)
	assertFloat(t, "skeletal muscle mass", result.SkeletalMuscleMass, 35.2)
	if result.SMMUnit != "kg" {
		t.Errorf("expected SMM unit 'kg', got %q", result.SMMUnit)
	}
	assertFloat(t, "BMR", result.BMR, 1847)
	assertFloat(t, "total body water", result.TotalBodyWater, 42.1)
	assertFloat(t, "lean body mass", result.LeanBodyMass, 66.4)

	if result.IsPartial() {
		t.Error("expected all fields to be extracted (not partial)")
	}
	if result.ExtractedFieldCount() != 6 {
		t.Errorf("expected 6 fields, got %d", result.ExtractedFieldCount())
	}
}

func TestInBodyExtractor_InBody270Format_ExtractsFields(t *testing.T) {
	text := `
InBody 270
Body Weight: 185.4 lbs
PBF: 22.3
SMM: 78.2 lbs
BMR: 1847
TBW: 42.1
LBM: 144.1 lbs
`
	ext := NewInBodyTextExtractor(plainTextExtractor)
	result, err := ext.ExtractInBody(context.Background(), []byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFloat(t, "weight", result.Weight, 185.4)
	if result.WeightUnit != "lbs" {
		t.Errorf("expected weight unit 'lbs', got %q", result.WeightUnit)
	}
	assertFloat(t, "body fat pct", result.BodyFatPct, 22.3)
	assertFloat(t, "SMM", result.SkeletalMuscleMass, 78.2)
	assertFloat(t, "BMR", result.BMR, 1847)
	assertFloat(t, "TBW", result.TotalBodyWater, 42.1)
	assertFloat(t, "LBM", result.LeanBodyMass, 144.1)

	if result.IsPartial() {
		t.Error("expected all fields to be extracted")
	}
}

func TestInBodyExtractor_InBody770Format_ExtractsFields(t *testing.T) {
	text := `
InBody 770

Body Composition
Weight (kg): 90.2
Skeletal Muscle Mass: 38.5 kg
Percent Body Fat: 18.7
Basal Metabolic Rate: 1920 kcal
Total Body Water: 45.3 L
Fat-Free Body Mass: 73.3 kg
`
	ext := NewInBodyTextExtractor(plainTextExtractor)
	result, err := ext.ExtractInBody(context.Background(), []byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFloat(t, "weight", result.Weight, 90.2)
	assertFloat(t, "body fat pct", result.BodyFatPct, 18.7)
	assertFloat(t, "SMM", result.SkeletalMuscleMass, 38.5)
	assertFloat(t, "BMR", result.BMR, 1920)
	assertFloat(t, "TBW", result.TotalBodyWater, 45.3)
	assertFloat(t, "LBM", result.LeanBodyMass, 73.3)
}

func TestInBodyExtractor_PartialExtraction_MissingBMR(t *testing.T) {
	text := `
Weight: 85.4 kg
Body Fat Percentage: 22.3
SMM: 35.2 kg
TBW: 42.1 L
LBM: 66.4 kg
`
	ext := NewInBodyTextExtractor(plainTextExtractor)
	result, err := ext.ExtractInBody(context.Background(), []byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFloat(t, "weight", result.Weight, 85.4)
	assertFloat(t, "body fat pct", result.BodyFatPct, 22.3)
	assertFloat(t, "SMM", result.SkeletalMuscleMass, 35.2)
	if result.BMR != nil {
		t.Error("expected BMR to be nil")
	}
	assertFloat(t, "TBW", result.TotalBodyWater, 42.1)
	assertFloat(t, "LBM", result.LeanBodyMass, 66.4)

	if !result.IsPartial() {
		t.Error("expected partial extraction with 5 fields")
	}
	if result.ExtractedFieldCount() != 5 {
		t.Errorf("expected 5 fields, got %d", result.ExtractedFieldCount())
	}
}

func TestInBodyExtractor_NoRecognizedFields_ReturnsEmptyResult(t *testing.T) {
	text := `
This is some random text that does not contain
any InBody metrics at all.
`
	ext := NewInBodyTextExtractor(plainTextExtractor)
	result, err := ext.ExtractInBody(context.Background(), []byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsEmpty() {
		t.Error("expected empty result for unrecognized content")
	}
}

func TestInBodyExtractor_LbsUnit_Recognized(t *testing.T) {
	text := `Weight: 185.4 lbs`
	ext := NewInBodyTextExtractor(plainTextExtractor)
	result, err := ext.ExtractInBody(context.Background(), []byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFloat(t, "weight", result.Weight, 185.4)
	if result.WeightUnit != "lbs" {
		t.Errorf("expected unit 'lbs', got %q", result.WeightUnit)
	}
}

func TestInBodyExtractor_FFMAlias_ExtractsLeanBodyMass(t *testing.T) {
	text := `FFM: 66.4 kg`
	ext := NewInBodyTextExtractor(plainTextExtractor)
	result, err := ext.ExtractInBody(context.Background(), []byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFloat(t, "lean body mass (FFM)", result.LeanBodyMass, 66.4)
}

func TestInBodyExtractor_WeightDecimalValues(t *testing.T) {
	text := `Weight: 72 kg`
	ext := NewInBodyTextExtractor(plainTextExtractor)
	result, err := ext.ExtractInBody(context.Background(), []byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFloat(t, "weight", result.Weight, 72.0)
}

func TestInBodyExtractor_BodyFatPctValidation_RejectsOver100(t *testing.T) {
	text := `Body Fat Percentage: 150`
	ext := NewInBodyTextExtractor(plainTextExtractor)
	result, err := ext.ExtractInBody(context.Background(), []byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.BodyFatPct != nil {
		t.Error("expected body fat pct to be nil for value > 100")
	}
}

// assertFloat is a test helper that checks a float pointer matches expected value.
func assertFloat(t *testing.T, name string, got *float64, expected float64) {
	t.Helper()
	if got == nil {
		t.Errorf("expected %s to be %.1f, got nil", name, expected)
		return
	}
	if math.Abs(*got-expected) > 0.01 {
		t.Errorf("expected %s to be %.1f, got %.1f", name, expected, *got)
	}
}
