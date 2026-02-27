package entities

import (
	"testing"
)

func TestNormalizeMarkerName_KnownMarkers_ReturnsType(t *testing.T) {
	tests := []struct {
		input    string
		expected LabMeasurementType
	}{
		{"Glucose, Fasting", LabTypeFastingGlucose},
		{"glucose", LabTypeFastingGlucose},
		{"Fasting Glucose", LabTypeFastingGlucose},
		{"Total Cholesterol", LabTypeTotalCholesterol},
		{"Cholesterol", LabTypeTotalCholesterol},
		{"LDL Cholesterol", LabTypeLDLCholesterol},
		{"LDL-C", LabTypeLDLCholesterol},
		{"LDL", LabTypeLDLCholesterol},
		{"HDL Cholesterol", LabTypeHDLCholesterol},
		{"HDL-C", LabTypeHDLCholesterol},
		{"HDL", LabTypeHDLCholesterol},
		{"Triglycerides", LabTypeTriglycerides},
		{"HbA1c", LabTypeHbA1c},
		{"Hemoglobin A1C", LabTypeHbA1c},
		{"A1C", LabTypeHbA1c},
		{"Testosterone", LabTypeTestosterone},
		{"Testosterone, Total", LabTypeTestosterone},
		{"TSH", LabTypeTSH},
		{"Vitamin D", LabTypeVitaminD},
		{"Vitamin D, 25-Hydroxy", LabTypeVitaminD},
		{"Iron", LabTypeIron},
		{"Iron, Serum", LabTypeIron},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeMarkerName(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeMarkerName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNormalizeMarkerName_UnknownMarker_ReturnsEmpty(t *testing.T) {
	got := NormalizeMarkerName("unknown marker xyz")
	if got != "" {
		t.Errorf("NormalizeMarkerName(unknown) = %q, want empty", got)
	}
}

func TestNormalizeMarkerName_ExtraWhitespace_Normalized(t *testing.T) {
	got := NormalizeMarkerName("  glucose   fasting  ")
	if got != LabTypeFastingGlucose {
		t.Errorf("NormalizeMarkerName with extra spaces = %q, want %q", got, LabTypeFastingGlucose)
	}
}

func TestDetermineFlag_Normal_ReturnsNormal(t *testing.T) {
	low := 70.0
	high := 100.0
	flag := DetermineFlag(85, &low, &high)
	if flag == nil || *flag != LabFlagNormal {
		t.Errorf("DetermineFlag(85, 70-100) = %v, want normal", flag)
	}
}

func TestDetermineFlag_Low_ReturnsLow(t *testing.T) {
	low := 70.0
	high := 100.0
	flag := DetermineFlag(65, &low, &high)
	if flag == nil || *flag != LabFlagLow {
		t.Errorf("DetermineFlag(65, 70-100) = %v, want low", flag)
	}
}

func TestDetermineFlag_High_ReturnsHigh(t *testing.T) {
	low := 70.0
	high := 100.0
	flag := DetermineFlag(105, &low, &high)
	if flag == nil || *flag != LabFlagHigh {
		t.Errorf("DetermineFlag(105, 70-100) = %v, want high", flag)
	}
}

func TestDetermineFlag_CriticalLow_ReturnsCriticalLow(t *testing.T) {
	low := 70.0
	high := 100.0
	// 50% below 70 = 35, so below 35 is critical
	flag := DetermineFlag(30, &low, &high)
	if flag == nil || *flag != LabFlagCriticalLow {
		t.Errorf("DetermineFlag(30, 70-100) = %v, want critical_low", flag)
	}
}

func TestDetermineFlag_CriticalHigh_ReturnsCriticalHigh(t *testing.T) {
	low := 70.0
	high := 100.0
	// 50% above 100 = 150, so above 150 is critical
	flag := DetermineFlag(160, &low, &high)
	if flag == nil || *flag != LabFlagCriticalHigh {
		t.Errorf("DetermineFlag(160, 70-100) = %v, want critical_high", flag)
	}
}

func TestDetermineFlag_NoRange_ReturnsNil(t *testing.T) {
	flag := DetermineFlag(85, nil, nil)
	if flag != nil {
		t.Errorf("DetermineFlag with no range = %v, want nil", flag)
	}
}

func TestDetermineFlag_OnlyHighBound_ReturnsFlag(t *testing.T) {
	high := 100.0
	flag := DetermineFlag(95, nil, &high)
	if flag == nil || *flag != LabFlagNormal {
		t.Errorf("DetermineFlag(95, nil-100) = %v, want normal", flag)
	}

	flag = DetermineFlag(110, nil, &high)
	if flag == nil || *flag != LabFlagHigh {
		t.Errorf("DetermineFlag(110, nil-100) = %v, want high", flag)
	}
}

func TestDetermineFlag_OnlyLowBound_ReturnsFlag(t *testing.T) {
	low := 40.0
	flag := DetermineFlag(50, &low, nil)
	if flag == nil || *flag != LabFlagNormal {
		t.Errorf("DetermineFlag(50, 40-nil) = %v, want normal", flag)
	}

	flag = DetermineFlag(35, &low, nil)
	if flag == nil || *flag != LabFlagLow {
		t.Errorf("DetermineFlag(35, 40-nil) = %v, want low", flag)
	}
}

func TestDetermineFlag_AtBoundary_ReturnsNormal(t *testing.T) {
	low := 70.0
	high := 100.0
	// Exactly at low boundary
	flag := DetermineFlag(70, &low, &high)
	if flag == nil || *flag != LabFlagNormal {
		t.Errorf("DetermineFlag(70, 70-100) = %v, want normal", flag)
	}

	// Exactly at high boundary
	flag = DetermineFlag(100, &low, &high)
	if flag == nil || *flag != LabFlagNormal {
		t.Errorf("DetermineFlag(100, 70-100) = %v, want normal", flag)
	}
}

func TestIsValidLabMeasurementType_Valid_ReturnsTrue(t *testing.T) {
	validTypes := []LabMeasurementType{
		LabTypeFastingGlucose, LabTypeTotalCholesterol, LabTypeLDLCholesterol,
		LabTypeHDLCholesterol, LabTypeTriglycerides, LabTypeHbA1c,
		LabTypeTestosterone, LabTypeTSH, LabTypeVitaminD, LabTypeIron,
	}
	for _, mt := range validTypes {
		if !IsValidLabMeasurementType(mt) {
			t.Errorf("IsValidLabMeasurementType(%q) = false, want true", mt)
		}
	}
}

func TestIsValidLabMeasurementType_Invalid_ReturnsFalse(t *testing.T) {
	if IsValidLabMeasurementType("unknown_type") {
		t.Error("IsValidLabMeasurementType(unknown_type) = true, want false")
	}
}

func TestRoundToThreeDecimals_Rounds(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{1.23456, 1.235},
		{1.0, 1.0},
		{98.1, 98.1},
		{0.0005, 0.001},
		{0.0004, 0.0},
	}
	for _, tt := range tests {
		got := RoundToThreeDecimals(tt.input)
		if got != tt.expected {
			t.Errorf("RoundToThreeDecimals(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestFloatPtr_ReturnsPointer(t *testing.T) {
	p := FloatPtr(42.5)
	if p == nil {
		t.Fatal("FloatPtr returned nil")
	}
	if *p != 42.5 {
		t.Errorf("FloatPtr(42.5) = %v, want 42.5", *p)
	}
}
