package inbody

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// mockTextExtract returns a function that returns the given text.
func mockTextExtract(text string) func([]byte) (string, error) {
	return func([]byte) (string, error) {
		return text, nil
	}
}

// mockTextExtractErr returns a function that returns an error.
func mockTextExtractErr(err error) func([]byte) (string, error) {
	return func([]byte) (string, error) {
		return "", err
	}
}

const inBody570Text = `InBody 570 Body Composition Analysis
Test Date: 01/15/2026
Name: John Doe
Gender: Male
Age: 30

Body Composition Analysis
Total Body Water: 42.1 L
Lean Body Mass: 66.5 kg
Body Weight: 85.5 kg

Muscle-Fat Analysis
Skeletal Muscle Mass: 35.2 kg
Body Fat Percentage: 22.3 %

Basal Metabolic Rate: 1847 kcal
`

const inBody270Text = `InBody 270
Date: 02/20/2026

Weight (kg): 78.0
PBF: 18.5
SMM: 32.1 kg
BMR (kcal): 1720
TBW (L): 39.8
LBM (kg): 63.6
`

const partialText = `InBody Report
Body Weight: 85.5 kg
Body Fat Percentage: 22.3 %
`

const noMetricsText = `This is some random text with no InBody data.
Hello world. Just filler text.
`

func assertFloat64Near(t *testing.T, label string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.01 {
		t.Errorf("%s: expected %.2f, got %.2f", label, want, got)
	}
}

func TestTextExtractor_ExtractInBody_InBody570_AllFields(t *testing.T) {
	ext := NewTextExtractor(mockTextExtract(inBody570Text))
	result, err := ext.ExtractInBody(context.Background(), []byte("fake-pdf"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Weight == nil {
		t.Fatal("expected weight to be extracted")
	}
	assertFloat64Near(t, "weight", *result.Weight, 85.5)
	if result.WeightUnit != "kg" {
		t.Errorf("expected weight unit 'kg', got %q", result.WeightUnit)
	}

	if result.BodyFatPct == nil {
		t.Fatal("expected body fat % to be extracted")
	}
	assertFloat64Near(t, "body_fat_pct", *result.BodyFatPct, 22.3)

	if result.SkeletalMuscleMass == nil {
		t.Fatal("expected SMM to be extracted")
	}
	assertFloat64Near(t, "smm", *result.SkeletalMuscleMass, 35.2)

	if result.BMR == nil {
		t.Fatal("expected BMR to be extracted")
	}
	assertFloat64Near(t, "bmr", *result.BMR, 1847.0)

	if result.TotalBodyWater == nil {
		t.Fatal("expected TBW to be extracted")
	}
	assertFloat64Near(t, "tbw", *result.TotalBodyWater, 42.1)

	if result.LeanBodyMass == nil {
		t.Fatal("expected LBM to be extracted")
	}
	assertFloat64Near(t, "lbm", *result.LeanBodyMass, 66.5)

	if result.IsPartial {
		t.Error("expected IsPartial to be false for complete extraction")
	}

	expectedDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	if !result.MeasuredAt.Equal(expectedDate) {
		t.Errorf("expected date %v, got %v", expectedDate, result.MeasuredAt)
	}
}

func TestTextExtractor_ExtractInBody_InBody270_AllFields(t *testing.T) {
	ext := NewTextExtractor(mockTextExtract(inBody270Text))
	result, err := ext.ExtractInBody(context.Background(), []byte("fake-pdf"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Weight == nil {
		t.Fatal("expected weight to be extracted")
	}
	assertFloat64Near(t, "weight", *result.Weight, 78.0)

	if result.BodyFatPct == nil {
		t.Fatal("expected body fat % to be extracted")
	}
	assertFloat64Near(t, "body_fat_pct", *result.BodyFatPct, 18.5)

	if result.SkeletalMuscleMass == nil {
		t.Fatal("expected SMM to be extracted")
	}
	assertFloat64Near(t, "smm", *result.SkeletalMuscleMass, 32.1)

	if result.BMR == nil {
		t.Fatal("expected BMR to be extracted")
	}
	assertFloat64Near(t, "bmr", *result.BMR, 1720.0)

	if result.TotalBodyWater == nil {
		t.Fatal("expected TBW to be extracted")
	}
	assertFloat64Near(t, "tbw", *result.TotalBodyWater, 39.8)

	if result.LeanBodyMass == nil {
		t.Fatal("expected LBM to be extracted")
	}
	assertFloat64Near(t, "lbm", *result.LeanBodyMass, 63.6)

	if result.IsPartial {
		t.Error("expected IsPartial to be false for complete extraction")
	}
}

func TestTextExtractor_ExtractInBody_Partial_SetsIsPartial(t *testing.T) {
	ext := NewTextExtractor(mockTextExtract(partialText))
	result, err := ext.ExtractInBody(context.Background(), []byte("fake-pdf"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.FieldCount() != 2 {
		t.Errorf("expected 2 fields, got %d", result.FieldCount())
	}
	if !result.IsPartial {
		t.Error("expected IsPartial to be true")
	}
	if result.Weight == nil {
		t.Fatal("expected weight to be extracted")
	}
	if result.BodyFatPct == nil {
		t.Fatal("expected body fat % to be extracted")
	}
}

func TestTextExtractor_ExtractInBody_EmptyPDF_ReturnsError(t *testing.T) {
	ext := NewTextExtractor(mockTextExtract(""))
	_, err := ext.ExtractInBody(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for empty PDF")
	}
}

func TestTextExtractor_ExtractInBody_TextExtractionError_ReturnsError(t *testing.T) {
	ext := NewTextExtractor(mockTextExtractErr(fmt.Errorf("corrupt PDF")))
	_, err := ext.ExtractInBody(context.Background(), []byte("bad-data"))
	if err == nil {
		t.Fatal("expected error for text extraction failure")
	}
}

func TestTextExtractor_ExtractInBody_NoMetrics_ReturnsError(t *testing.T) {
	ext := NewTextExtractor(mockTextExtract(noMetricsText))
	_, err := ext.ExtractInBody(context.Background(), []byte("fake-pdf"))
	if err == nil {
		t.Fatal("expected error when no metrics found")
	}
}

func TestTextExtractor_ExtractInBody_EmptyText_ReturnsError(t *testing.T) {
	ext := NewTextExtractor(mockTextExtract(""))
	_, err := ext.ExtractInBody(context.Background(), []byte("fake-pdf"))
	if err == nil {
		t.Fatal("expected error for empty text")
	}
}

func TestTextExtractor_ExtractInBody_LbsUnit_NormalizesCorrectly(t *testing.T) {
	text := `Body Weight: 185.4 lbs
Skeletal Muscle Mass: 78.2 lbs
Lean Body Mass: 144.1 lbs
Body Fat Percentage: 22.0 %`

	ext := NewTextExtractor(mockTextExtract(text))
	result, err := ext.ExtractInBody(context.Background(), []byte("fake-pdf"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.WeightUnit != "lbs" {
		t.Errorf("expected 'lbs', got %q", result.WeightUnit)
	}
	if result.SkeletalMuscleUnit != "lbs" {
		t.Errorf("expected 'lbs', got %q", result.SkeletalMuscleUnit)
	}
	if result.LeanBodyMassUnit != "lbs" {
		t.Errorf("expected 'lbs', got %q", result.LeanBodyMassUnit)
	}
}

func TestTextExtractor_ExtractInBody_ISODate_ParsesCorrectly(t *testing.T) {
	text := `Date: 2026-03-15
Weight: 80.0 kg
Body Fat: 20.0 %`

	ext := NewTextExtractor(mockTextExtract(text))
	result, err := ext.ExtractInBody(context.Background(), []byte("fake-pdf"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The YYYY-MM-DD pattern is matched by the second datePattern
	expectedDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	if !result.MeasuredAt.Equal(expectedDate) {
		t.Errorf("expected date %v, got %v", expectedDate, result.MeasuredAt)
	}
}

func TestTextExtractor_ExtractInBody_ToMeasurements_CorrectConversion(t *testing.T) {
	ext := NewTextExtractor(mockTextExtract(inBody570Text))
	result, err := ext.ExtractInBody(context.Background(), []byte("fake-pdf"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	measurements := result.ToMeasurements("client-1", "coach-1", "artifact-1")
	if len(measurements) != 6 {
		t.Fatalf("expected 6 measurements, got %d", len(measurements))
	}

	// Verify measurement types
	typeMap := make(map[entities.MeasurementType]bool)
	for _, m := range measurements {
		typeMap[m.MeasurementType] = true
	}
	expectedTypes := []entities.MeasurementType{
		entities.MeasurementTypeWeight,
		entities.MeasurementTypeBodyFatPct,
		entities.MeasurementTypeSkeletalMuscleMass,
		entities.MeasurementTypeBMR,
		entities.MeasurementTypeTotalBodyWater,
		entities.MeasurementTypeLeanBodyMass,
	}
	for _, et := range expectedTypes {
		if !typeMap[et] {
			t.Errorf("missing measurement type: %s", et)
		}
	}
}

func TestExtractWeight_VariousFormats(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantVal  float64
		wantUnit string
	}{
		{"standard", "Body Weight: 85.5 kg", 85.5, "kg"},
		{"no colon", "Weight 78.0 kg", 78.0, "kg"},
		{"lbs", "Weight: 185.4 lbs", 185.4, "lbs"},
		{"parenthesized unit", "Weight (kg): 78.0", 78.0, "kg"},
		{"equals sign", "Weight = 80.0 kg", 80.0, "kg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, unit := extractWeight(tt.text)
			if val == nil {
				t.Fatal("expected value")
			}
			assertFloat64Near(t, "weight", *val, tt.wantVal)
			if unit != tt.wantUnit {
				t.Errorf("expected unit %q, got %q", tt.wantUnit, unit)
			}
		})
	}
}

func TestExtractBodyFatPct_VariousFormats(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantVal float64
	}{
		{"standard", "Body Fat Percentage: 22.3", 22.3},
		{"with percent sign", "Body Fat: 22.3 %", 22.3},
		{"PBF format", "PBF: 18.5", 18.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := extractBodyFatPct(tt.text)
			if val == nil {
				t.Fatal("expected value")
			}
			assertFloat64Near(t, "body_fat_pct", *val, tt.wantVal)
		})
	}
}

func TestExtractBMR_VariousFormats(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantVal float64
	}{
		{"standard", "Basal Metabolic Rate: 1847 kcal", 1847.0},
		{"BMR shorthand", "BMR: 1720", 1720.0},
		{"BMR with kcal paren", "BMR (kcal): 1600", 1600.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := extractBMR(tt.text)
			if val == nil {
				t.Fatal("expected value")
			}
			assertFloat64Near(t, "bmr", *val, tt.wantVal)
		})
	}
}

func TestExtractTotalBodyWater_VariousFormats(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantVal float64
	}{
		{"standard", "Total Body Water: 42.1 L", 42.1},
		{"TBW shorthand", "TBW: 39.8", 39.8},
		{"TBW with L paren", "TBW (L): 40.5", 40.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := extractTotalBodyWater(tt.text)
			if val == nil {
				t.Fatal("expected value")
			}
			assertFloat64Near(t, "tbw", *val, tt.wantVal)
		})
	}
}

func TestExtractSkeletalMuscleMass_VariousFormats(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantVal  float64
		wantUnit string
	}{
		{"standard", "Skeletal Muscle Mass: 35.2 kg", 35.2, "kg"},
		{"SMM shorthand", "SMM: 32.1 kg", 32.1, "kg"},
		{"SMM parenthesized", "SMM (kg): 30.0", 30.0, "kg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, unit := extractSkeletalMuscleMass(tt.text)
			if val == nil {
				t.Fatal("expected value")
			}
			assertFloat64Near(t, "smm", *val, tt.wantVal)
			if unit != tt.wantUnit {
				t.Errorf("expected unit %q, got %q", tt.wantUnit, unit)
			}
		})
	}
}

func TestExtractLeanBodyMass_VariousFormats(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantVal  float64
		wantUnit string
	}{
		{"standard", "Lean Body Mass: 66.5 kg", 66.5, "kg"},
		{"LBM shorthand", "LBM: 63.6 kg", 63.6, "kg"},
		{"fat free mass", "Fat-Free Mass: 60.0 kg", 60.0, "kg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, unit := extractLeanBodyMass(tt.text)
			if val == nil {
				t.Fatal("expected value")
			}
			assertFloat64Near(t, "lbm", *val, tt.wantVal)
			if unit != tt.wantUnit {
				t.Errorf("expected unit %q, got %q", tt.wantUnit, unit)
			}
		})
	}
}

func TestExtractDate_VariousFormats(t *testing.T) {
	tests := []struct {
		name string
		text string
		want time.Time
	}{
		{"MM/DD/YYYY", "Test Date: 01/15/2026", time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)},
		{"YYYY-MM-DD", "2026-03-15", time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)},
		{"with dashes", "Date: 02-20-2026", time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDate(tt.text)
			if !got.Equal(tt.want) {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestExtractWeight_NoMatch_ReturnsNil(t *testing.T) {
	val, unit := extractWeight("no weight here")
	if val != nil {
		t.Errorf("expected nil, got %v", *val)
	}
	if unit != "" {
		t.Errorf("expected empty unit, got %q", unit)
	}
}

func TestExtractBodyFatPct_NoMatch_ReturnsNil(t *testing.T) {
	val := extractBodyFatPct("no body fat here")
	if val != nil {
		t.Errorf("expected nil, got %v", *val)
	}
}

func TestExtractBMR_NoMatch_ReturnsNil(t *testing.T) {
	val := extractBMR("no BMR here")
	if val != nil {
		t.Errorf("expected nil, got %v", *val)
	}
}

func TestExtractTotalBodyWater_NoMatch_ReturnsNil(t *testing.T) {
	val := extractTotalBodyWater("no TBW here")
	if val != nil {
		t.Errorf("expected nil, got %v", *val)
	}
}

func TestExtractSkeletalMuscleMass_NoMatch_ReturnsNil(t *testing.T) {
	val, unit := extractSkeletalMuscleMass("no SMM here")
	if val != nil {
		t.Errorf("expected nil, got %v", *val)
	}
	if unit != "" {
		t.Errorf("expected empty unit, got %q", unit)
	}
}

func TestExtractLeanBodyMass_NoMatch_ReturnsNil(t *testing.T) {
	val, unit := extractLeanBodyMass("no LBM here")
	if val != nil {
		t.Errorf("expected nil, got %v", *val)
	}
	if unit != "" {
		t.Errorf("expected empty unit, got %q", unit)
	}
}
