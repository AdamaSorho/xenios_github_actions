package pdf

import (
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

const sampleInBody570Text = `InBody 570 Body Composition Analysis
Date: 2026-01-15
ID: 12345

Body Composition Analysis
Weight: 85.4 kg
Skeletal Muscle Mass: 38.2 kg
Body Fat Percentage: 22.3%
Basal Metabolic Rate: 1847 kcal
Total Body Water: 42.1 L
Lean Body Mass: 66.4 kg

Segmental Lean Analysis
Right Arm: 3.5 kg
Left Arm: 3.4 kg
Trunk: 28.1 kg
Right Leg: 10.2 kg
Left Leg: 10.1 kg
`

const sampleInBody270Text = `InBody 270 Result Sheet

Body Weight - 185.4 lbs
PBF - 25.1 %
SMM - 78.2 lbs
BMR - 1920 kcal
TBW - 38.5 L
Fat Free Mass: 139.0 lbs
`

const partialInBodyText = `InBody Analysis Report
Weight: 90.0 kg
Body Fat Percentage: 18.5%
Skeletal Muscle Mass: 42.0 kg
`

const emptyText = ``

const corruptText = `This is not an InBody scan at all.
Random numbers and text without any recognizable fields.
`

func TestParseInBodyText_InBody570_ExtractsAllFields(t *testing.T) {
	now := time.Now()
	result := ParseInBodyText(sampleInBody570Text, "client-1", "coach-1", "artifact-1", now)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Measurements) != 6 {
		t.Fatalf("expected 6 measurements, got %d", len(result.Measurements))
	}
	if result.IsPartial {
		t.Error("expected complete extraction")
	}

	expected := map[entities.MeasurementType]struct {
		value float64
		unit  string
	}{
		entities.MeasurementTypeWeight:             {85.4, "kg"},
		entities.MeasurementTypeBodyFatPct:         {22.3, "%"},
		entities.MeasurementTypeSkeletalMuscleMass: {38.2, "kg"},
		entities.MeasurementTypeBMR:                {1847, "kcal"},
		entities.MeasurementTypeTotalBodyWater:     {42.1, "L"},
		entities.MeasurementTypeLeanBodyMass:       {66.4, "kg"},
	}

	for _, m := range result.Measurements {
		exp, ok := expected[m.MeasurementType]
		if !ok {
			t.Errorf("unexpected measurement type: %s", m.MeasurementType)
			continue
		}
		if m.Value != exp.value {
			t.Errorf("type %s: expected value %f, got %f", m.MeasurementType, exp.value, m.Value)
		}
		if m.Unit != exp.unit {
			t.Errorf("type %s: expected unit %q, got %q", m.MeasurementType, exp.unit, m.Unit)
		}
		if m.ClientID != "client-1" {
			t.Errorf("expected client_id 'client-1', got %q", m.ClientID)
		}
		if m.RecordedBy != "coach-1" {
			t.Errorf("expected recorded_by 'coach-1', got %q", m.RecordedBy)
		}
		if m.ArtifactID != "artifact-1" {
			t.Errorf("expected artifact_id 'artifact-1', got %q", m.ArtifactID)
		}
	}
}

func TestParseInBodyText_InBody270_ExtractsAllFields(t *testing.T) {
	now := time.Now()
	result := ParseInBodyText(sampleInBody270Text, "client-2", "coach-2", "artifact-2", now)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Measurements) != 6 {
		t.Fatalf("expected 6 measurements, got %d", len(result.Measurements))
	}
	if result.IsPartial {
		t.Error("expected complete extraction")
	}

	// Check weight was extracted with lbs
	for _, m := range result.Measurements {
		if m.MeasurementType == entities.MeasurementTypeWeight {
			if m.Value != 185.4 {
				t.Errorf("expected weight 185.4, got %f", m.Value)
			}
			if m.Unit != "lbs" {
				t.Errorf("expected unit lbs, got %q", m.Unit)
			}
		}
	}
}

func TestParseInBodyText_PartialExtraction_SetsIsPartial(t *testing.T) {
	now := time.Now()
	result := ParseInBodyText(partialInBodyText, "client-1", "coach-1", "artifact-1", now)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Measurements) != 3 {
		t.Fatalf("expected 3 measurements, got %d", len(result.Measurements))
	}
	if !result.IsPartial {
		t.Error("expected partial extraction")
	}
	if len(result.Errors) == 0 {
		t.Error("expected errors for missing fields")
	}

	// All measurements should have partial status
	for _, m := range result.Measurements {
		if m.ExtractionStatus != entities.ExtractionStatusPartial {
			t.Errorf("expected partial status, got %s", m.ExtractionStatus)
		}
	}
}

func TestParseInBodyText_EmptyText_ReturnsNoMeasurements(t *testing.T) {
	now := time.Now()
	result := ParseInBodyText(emptyText, "client-1", "coach-1", "artifact-1", now)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Measurements) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(result.Measurements))
	}
	if len(result.Errors) != 6 {
		t.Errorf("expected 6 errors for all missing fields, got %d", len(result.Errors))
	}
}

func TestParseInBodyText_CorruptText_ReturnsNoMeasurements(t *testing.T) {
	now := time.Now()
	result := ParseInBodyText(corruptText, "client-1", "coach-1", "artifact-1", now)

	if len(result.Measurements) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(result.Measurements))
	}
}

func TestParseInBodyText_MeasuredAtPropagated(t *testing.T) {
	fixedTime := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	result := ParseInBodyText(sampleInBody570Text, "client-1", "coach-1", "artifact-1", fixedTime)

	for _, m := range result.Measurements {
		if !m.MeasuredAt.Equal(fixedTime) {
			t.Errorf("expected measured_at %v, got %v", fixedTime, m.MeasuredAt)
		}
	}
}

func TestNormalizeUnit_Variants(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"kg", "kg"},
		{"lbs", "lbs"},
		{"lb", "lbs"},
		{"L", "L"},
		{"liter", "L"},
		{"kcal", "kcal"},
		{" kg ", "kg"},
		{"LBS", "lbs"},
	}

	for _, tc := range tests {
		got := normalizeUnit(tc.input)
		if got != tc.expected {
			t.Errorf("normalizeUnit(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestMatchField_NoMatch_ReturnsFalse(t *testing.T) {
	field := inbodyFields[0] // weight
	_, _, found := matchField("no weight data here", field)
	if found {
		t.Error("expected no match")
	}
}

func TestMatchField_ValidMatch_ReturnsValue(t *testing.T) {
	field := inbodyFields[0] // weight
	value, unit, found := matchField("Weight: 85.4 kg", field)
	if !found {
		t.Fatal("expected match")
	}
	if value != 85.4 {
		t.Errorf("expected 85.4, got %f", value)
	}
	if unit != "kg" {
		t.Errorf("expected 'kg', got %q", unit)
	}
}
