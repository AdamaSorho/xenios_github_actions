package entities

import "testing"

func TestIsValidWearableSource_ValidSources_ReturnsTrue(t *testing.T) {
	sources := []WearableSource{
		WearableSourceWhoop,
		WearableSourceGarmin,
		WearableSourceAppleHealth,
		WearableSourceOura,
		WearableSourceFitbit,
	}
	for _, src := range sources {
		t.Run(string(src), func(t *testing.T) {
			if !IsValidWearableSource(src) {
				t.Errorf("expected %q to be valid", src)
			}
		})
	}
}

func TestIsValidWearableSource_InvalidSource_ReturnsFalse(t *testing.T) {
	if IsValidWearableSource("nonexistent") {
		t.Error("expected invalid source to return false")
	}
}

func TestIsValidMeasurementType_ValidTypes_ReturnsTrue(t *testing.T) {
	types := []MeasurementType{
		MeasurementTypeHRV,
		MeasurementTypeSleepDuration,
		MeasurementTypeRecovery,
		MeasurementTypeStrain,
		MeasurementTypeRestingHR,
		MeasurementTypeSteps,
		MeasurementTypeSleepQuality,
	}
	for _, mt := range types {
		t.Run(string(mt), func(t *testing.T) {
			if !IsValidMeasurementType(mt) {
				t.Errorf("expected %q to be valid", mt)
			}
		})
	}
}

func TestIsValidMeasurementType_InvalidType_ReturnsFalse(t *testing.T) {
	if IsValidMeasurementType("nonexistent") {
		t.Error("expected invalid type to return false")
	}
}

func TestUnitForMeasurementType_KnownTypes_ReturnsUnit(t *testing.T) {
	tests := []struct {
		mt   MeasurementType
		unit string
	}{
		{MeasurementTypeHRV, "ms"},
		{MeasurementTypeSleepDuration, "hrs"},
		{MeasurementTypeRecovery, "score"},
		{MeasurementTypeStrain, "score"},
		{MeasurementTypeRestingHR, "bpm"},
		{MeasurementTypeSteps, "count"},
		{MeasurementTypeSleepQuality, "score"},
	}
	for _, tt := range tests {
		t.Run(string(tt.mt), func(t *testing.T) {
			if got := UnitForMeasurementType(tt.mt); got != tt.unit {
				t.Errorf("expected unit %q, got %q", tt.unit, got)
			}
		})
	}
}

func TestUnitForMeasurementType_UnknownType_ReturnsEmpty(t *testing.T) {
	if got := UnitForMeasurementType("unknown"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
