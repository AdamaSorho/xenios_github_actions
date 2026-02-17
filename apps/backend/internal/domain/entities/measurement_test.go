package entities

import (
	"testing"
)

func TestIsValidMeasurementFlag_ValidFlags_ReturnsTrue(t *testing.T) {
	validFlags := []MeasurementFlag{FlagNormal, FlagLow, FlagHigh, FlagCriticalLow, FlagCriticalHigh}
	for _, f := range validFlags {
		if !IsValidMeasurementFlag(f) {
			t.Errorf("expected %q to be valid", f)
		}
	}
}

func TestIsValidMeasurementFlag_InvalidFlag_ReturnsFalse(t *testing.T) {
	if IsValidMeasurementFlag("unknown") {
		t.Error("expected 'unknown' flag to be invalid")
	}
	if IsValidMeasurementFlag("") {
		t.Error("expected empty flag to be invalid")
	}
}

func TestDetermineFlag_BothBounds_Normal(t *testing.T) {
	low := 70.0
	high := 100.0
	flag := DetermineFlag(85, &low, &high)
	if flag != FlagNormal {
		t.Errorf("expected normal, got %q", flag)
	}
}

func TestDetermineFlag_BothBounds_Low(t *testing.T) {
	low := 70.0
	high := 100.0
	flag := DetermineFlag(50, &low, &high)
	if flag != FlagLow {
		t.Errorf("expected low, got %q", flag)
	}
}

func TestDetermineFlag_BothBounds_High(t *testing.T) {
	low := 70.0
	high := 100.0
	flag := DetermineFlag(120, &low, &high)
	if flag != FlagHigh {
		t.Errorf("expected high, got %q", flag)
	}
}

func TestDetermineFlag_BothBounds_AtLowBoundary_Normal(t *testing.T) {
	low := 70.0
	high := 100.0
	flag := DetermineFlag(70, &low, &high)
	if flag != FlagNormal {
		t.Errorf("expected normal at lower boundary, got %q", flag)
	}
}

func TestDetermineFlag_BothBounds_AtHighBoundary_Normal(t *testing.T) {
	low := 70.0
	high := 100.0
	flag := DetermineFlag(100, &low, &high)
	if flag != FlagNormal {
		t.Errorf("expected normal at upper boundary, got %q", flag)
	}
}

func TestDetermineFlag_UpperBoundOnly_Normal(t *testing.T) {
	high := 100.0
	flag := DetermineFlag(80, nil, &high)
	if flag != FlagNormal {
		t.Errorf("expected normal, got %q", flag)
	}
}

func TestDetermineFlag_UpperBoundOnly_High(t *testing.T) {
	high := 100.0
	flag := DetermineFlag(142, nil, &high)
	if flag != FlagHigh {
		t.Errorf("expected high, got %q", flag)
	}
}

func TestDetermineFlag_LowerBoundOnly_Normal(t *testing.T) {
	low := 40.0
	flag := DetermineFlag(55, &low, nil)
	if flag != FlagNormal {
		t.Errorf("expected normal, got %q", flag)
	}
}

func TestDetermineFlag_LowerBoundOnly_Low(t *testing.T) {
	low := 40.0
	flag := DetermineFlag(30, &low, nil)
	if flag != FlagLow {
		t.Errorf("expected low, got %q", flag)
	}
}

func TestDetermineFlag_NoBounds_Empty(t *testing.T) {
	flag := DetermineFlag(85, nil, nil)
	if flag != "" {
		t.Errorf("expected empty flag when no bounds, got %q", flag)
	}
}

func TestKnownLabMarkers_CommonNames(t *testing.T) {
	tests := []struct {
		name     string
		expected LabMeasurementType
	}{
		{"glucose", LabFastingGlucose},
		{"glucose, fasting", LabFastingGlucose},
		{"fasting glucose", LabFastingGlucose},
		{"ldl cholesterol", LabLDLCholesterol},
		{"ldl-c", LabLDLCholesterol},
		{"hdl cholesterol", LabHDLCholesterol},
		{"total cholesterol", LabTotalCholesterol},
		{"triglycerides", LabTriglycerides},
		{"hba1c", LabHbA1c},
		{"hemoglobin a1c", LabHbA1c},
		{"a1c", LabHbA1c},
		{"testosterone", LabTestosterone},
		{"tsh", LabTSH},
		{"vitamin d", LabVitaminD},
		{"iron", LabIron},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := KnownLabMarkers[tt.name]
			if !ok {
				t.Fatalf("marker %q not found in KnownLabMarkers", tt.name)
			}
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
