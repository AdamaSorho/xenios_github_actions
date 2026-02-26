package entities

import "testing"

func TestIsValidMeasurementType_KnownType_ReturnsTrue(t *testing.T) {
	knownTypes := []string{"weight", "body_fat_pct", "skeletal_muscle_mass", "ldl_cholesterol", "calories", "protein"}
	for _, mt := range knownTypes {
		if !IsValidMeasurementType(mt) {
			t.Errorf("expected %q to be a valid measurement type", mt)
		}
	}
}

func TestIsValidMeasurementType_UnknownType_ReturnsFalse(t *testing.T) {
	if IsValidMeasurementType("unknown_metric") {
		t.Error("expected unknown_metric to be invalid")
	}
	if IsValidMeasurementType("") {
		t.Error("expected empty string to be invalid")
	}
}

func TestIsValidMeasurementFlag_KnownFlags_ReturnsTrue(t *testing.T) {
	flags := []string{"low", "high", "critical", "normal"}
	for _, f := range flags {
		if !IsValidMeasurementFlag(f) {
			t.Errorf("expected %q to be a valid flag", f)
		}
	}
}

func TestIsValidMeasurementFlag_UnknownFlag_ReturnsFalse(t *testing.T) {
	if IsValidMeasurementFlag("unknown") {
		t.Error("expected unknown to be invalid")
	}
	if IsValidMeasurementFlag("") {
		t.Error("expected empty string to be invalid")
	}
}
