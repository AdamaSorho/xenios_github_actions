package entities

import (
	"testing"
	"time"
)

func TestIsValidMeasurementType_ValidTypes_ReturnsTrue(t *testing.T) {
	validTypes := []MeasurementType{
		MeasurementTypeWeight,
		MeasurementTypeBodyFatPct,
		MeasurementTypeSkeletalMuscleMass,
		MeasurementTypeBMR,
		MeasurementTypeTotalBodyWater,
		MeasurementTypeLeanBodyMass,
	}
	for _, mt := range validTypes {
		t.Run(string(mt), func(t *testing.T) {
			if !IsValidMeasurementType(mt) {
				t.Errorf("expected %q to be valid", mt)
			}
		})
	}
}

func TestIsValidMeasurementType_InvalidType_ReturnsFalse(t *testing.T) {
	if IsValidMeasurementType("unknown_type") {
		t.Error("expected 'unknown_type' to be invalid")
	}
}

func TestValidateMeasurement_ValidInput_ReturnsNil(t *testing.T) {
	m := &Measurement{
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: MeasurementTypeWeight,
		Value:           85.5,
		Unit:            "kg",
		MeasuredAt:      time.Now(),
	}
	if err := ValidateMeasurement(m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateMeasurement_EmptyClientID_ReturnsError(t *testing.T) {
	m := &Measurement{
		RecordedBy:      "coach-1",
		MeasurementType: MeasurementTypeWeight,
		Value:           85.5,
		Unit:            "kg",
		MeasuredAt:      time.Now(),
	}
	err := ValidateMeasurement(m)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "client_id is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateMeasurement_EmptyRecordedBy_ReturnsError(t *testing.T) {
	m := &Measurement{
		ClientID:        "client-1",
		MeasurementType: MeasurementTypeWeight,
		Value:           85.5,
		Unit:            "kg",
		MeasuredAt:      time.Now(),
	}
	err := ValidateMeasurement(m)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "recorded_by is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateMeasurement_InvalidType_ReturnsError(t *testing.T) {
	m := &Measurement{
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "bogus",
		Value:           85.5,
		Unit:            "kg",
		MeasuredAt:      time.Now(),
	}
	err := ValidateMeasurement(m)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateMeasurement_EmptyUnit_ReturnsError(t *testing.T) {
	m := &Measurement{
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: MeasurementTypeWeight,
		Value:           85.5,
		MeasuredAt:      time.Now(),
	}
	err := ValidateMeasurement(m)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "unit is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateMeasurement_ZeroMeasuredAt_ReturnsError(t *testing.T) {
	m := &Measurement{
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: MeasurementTypeWeight,
		Value:           85.5,
		Unit:            "kg",
	}
	err := ValidateMeasurement(m)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "measured_at is required" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInBodyResult_FieldCount_AllFields_Returns6(t *testing.T) {
	w := 85.5
	bf := 22.3
	smm := 35.2
	bmr := 1847.0
	tbw := 42.1
	lbm := 66.5

	r := &InBodyResult{
		Weight:             &w,
		BodyFatPct:         &bf,
		SkeletalMuscleMass: &smm,
		BMR:                &bmr,
		TotalBodyWater:     &tbw,
		LeanBodyMass:       &lbm,
	}
	if r.FieldCount() != 6 {
		t.Errorf("expected 6, got %d", r.FieldCount())
	}
}

func TestInBodyResult_FieldCount_Partial_ReturnsCorrectCount(t *testing.T) {
	w := 85.5
	bf := 22.3
	r := &InBodyResult{
		Weight:     &w,
		BodyFatPct: &bf,
	}
	if r.FieldCount() != 2 {
		t.Errorf("expected 2, got %d", r.FieldCount())
	}
}

func TestInBodyResult_FieldCount_Empty_ReturnsZero(t *testing.T) {
	r := &InBodyResult{}
	if r.FieldCount() != 0 {
		t.Errorf("expected 0, got %d", r.FieldCount())
	}
}

func TestInBodyResult_ToMeasurements_AllFields_Returns6Measurements(t *testing.T) {
	w := 85.5
	bf := 22.3
	smm := 35.2
	bmr := 1847.0
	tbw := 42.1
	lbm := 66.5
	now := time.Now()

	r := &InBodyResult{
		Weight:             &w,
		WeightUnit:         "kg",
		BodyFatPct:         &bf,
		SkeletalMuscleMass: &smm,
		SkeletalMuscleUnit: "kg",
		BMR:                &bmr,
		TotalBodyWater:     &tbw,
		LeanBodyMass:       &lbm,
		LeanBodyMassUnit:   "kg",
		MeasuredAt:         now,
	}
	measurements := r.ToMeasurements("client-1", "coach-1", "artifact-1")
	if len(measurements) != 6 {
		t.Fatalf("expected 6 measurements, got %d", len(measurements))
	}

	// Verify all measurements have correct client/coach/artifact
	for _, m := range measurements {
		if m.ClientID != "client-1" {
			t.Errorf("expected client_id 'client-1', got %q", m.ClientID)
		}
		if m.RecordedBy != "coach-1" {
			t.Errorf("expected recorded_by 'coach-1', got %q", m.RecordedBy)
		}
		if m.ArtifactID != "artifact-1" {
			t.Errorf("expected artifact_id 'artifact-1', got %q", m.ArtifactID)
		}
		if m.MeasuredAt != now {
			t.Errorf("expected measured_at %v, got %v", now, m.MeasuredAt)
		}
	}

	// Verify first measurement is weight
	if measurements[0].MeasurementType != MeasurementTypeWeight {
		t.Errorf("expected weight, got %q", measurements[0].MeasurementType)
	}
	if measurements[0].Value != 85.5 {
		t.Errorf("expected 85.5, got %f", measurements[0].Value)
	}
	if measurements[0].Unit != "kg" {
		t.Errorf("expected 'kg', got %q", measurements[0].Unit)
	}
}

func TestInBodyResult_ToMeasurements_PartialFields_ReturnsOnlyExtracted(t *testing.T) {
	w := 85.5
	r := &InBodyResult{
		Weight:     &w,
		WeightUnit: "kg",
		MeasuredAt: time.Now(),
	}
	measurements := r.ToMeasurements("client-1", "coach-1", "artifact-1")
	if len(measurements) != 1 {
		t.Fatalf("expected 1 measurement, got %d", len(measurements))
	}
	if measurements[0].MeasurementType != MeasurementTypeWeight {
		t.Errorf("expected weight, got %q", measurements[0].MeasurementType)
	}
}

func TestInBodyResult_ToMeasurements_NoFields_ReturnsEmpty(t *testing.T) {
	r := &InBodyResult{MeasuredAt: time.Now()}
	measurements := r.ToMeasurements("client-1", "coach-1", "artifact-1")
	if len(measurements) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(measurements))
	}
}
