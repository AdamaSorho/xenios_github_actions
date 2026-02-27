package entities

import "testing"

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
		if !IsValidMeasurementType(mt) {
			t.Errorf("expected %q to be valid", mt)
		}
	}
}

func TestIsValidMeasurementType_InvalidType_ReturnsFalse(t *testing.T) {
	if IsValidMeasurementType("unknown_type") {
		t.Error("expected 'unknown_type' to be invalid")
	}
}

func TestIsValidMeasurementType_EmptyString_ReturnsFalse(t *testing.T) {
	if IsValidMeasurementType("") {
		t.Error("expected empty string to be invalid")
	}
}
