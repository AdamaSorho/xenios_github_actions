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
			t.Errorf("expected %q to be a valid measurement type", mt)
		}
	}
}

func TestIsValidMeasurementType_InvalidType_ReturnsFalse(t *testing.T) {
	invalid := MeasurementType("unknown_type")
	if IsValidMeasurementType(invalid) {
		t.Errorf("expected %q to be invalid", invalid)
	}
}
