package entities

import "testing"

func TestIsValidMeasurementType_ValidTypes(t *testing.T) {
	validTypes := []MeasurementType{
		MeasurementTypeWeight,
		MeasurementTypeBodyFatPct,
		MeasurementTypeSkeletalMuscleMass,
		MeasurementTypeBMR,
		MeasurementTypeTotalBodyWater,
		MeasurementTypeLeanBodyMass,
	}
	for _, mt := range validTypes {
		if !IsValidMeasurementType(mt) {
			t.Errorf("expected %q to be a valid measurement type", mt)
		}
	}
}

func TestIsValidMeasurementType_InvalidType(t *testing.T) {
	if IsValidMeasurementType("unknown_type") {
		t.Error("expected 'unknown_type' to be invalid")
	}
}

func TestInBodyResult_ExtractedFieldCount_AllFields(t *testing.T) {
	w, bf, smm, bmr, tbw, lbm := 85.0, 22.3, 35.0, 1847.0, 42.0, 65.0
	r := &InBodyResult{
		Weight:             &w,
		BodyFatPct:         &bf,
		SkeletalMuscleMass: &smm,
		BMR:                &bmr,
		TotalBodyWater:     &tbw,
		LeanBodyMass:       &lbm,
	}
	if r.ExtractedFieldCount() != 6 {
		t.Errorf("expected 6 fields, got %d", r.ExtractedFieldCount())
	}
}

func TestInBodyResult_ExtractedFieldCount_PartialFields(t *testing.T) {
	w := 85.0
	bf := 22.3
	r := &InBodyResult{
		Weight:     &w,
		BodyFatPct: &bf,
	}
	if r.ExtractedFieldCount() != 2 {
		t.Errorf("expected 2 fields, got %d", r.ExtractedFieldCount())
	}
}

func TestInBodyResult_ExtractedFieldCount_NoFields(t *testing.T) {
	r := &InBodyResult{}
	if r.ExtractedFieldCount() != 0 {
		t.Errorf("expected 0 fields, got %d", r.ExtractedFieldCount())
	}
}

func TestInBodyResult_IsPartial_True(t *testing.T) {
	w := 85.0
	r := &InBodyResult{Weight: &w}
	if !r.IsPartial() {
		t.Error("expected IsPartial() to be true with 1 field")
	}
}

func TestInBodyResult_IsPartial_False(t *testing.T) {
	w, bf, smm, bmr, tbw, lbm := 85.0, 22.3, 35.0, 1847.0, 42.0, 65.0
	r := &InBodyResult{
		Weight:             &w,
		BodyFatPct:         &bf,
		SkeletalMuscleMass: &smm,
		BMR:                &bmr,
		TotalBodyWater:     &tbw,
		LeanBodyMass:       &lbm,
	}
	if r.IsPartial() {
		t.Error("expected IsPartial() to be false with all 6 fields")
	}
}

func TestInBodyResult_IsEmpty_True(t *testing.T) {
	r := &InBodyResult{}
	if !r.IsEmpty() {
		t.Error("expected IsEmpty() to be true")
	}
}

func TestInBodyResult_IsEmpty_False(t *testing.T) {
	w := 85.0
	r := &InBodyResult{Weight: &w}
	if r.IsEmpty() {
		t.Error("expected IsEmpty() to be false with weight set")
	}
}
