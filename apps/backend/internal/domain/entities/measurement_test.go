package entities

import (
	"testing"
)

func TestIsValidMeasurementFlag_ValidFlags_ReturnsTrue(t *testing.T) {
	flags := []MeasurementFlag{FlagNormal, FlagLow, FlagHigh, FlagCriticalLow, FlagCriticalHigh}
	for _, f := range flags {
		if !IsValidMeasurementFlag(f) {
			t.Errorf("expected %q to be valid", f)
		}
	}
}

func TestIsValidMeasurementFlag_InvalidFlag_ReturnsFalse(t *testing.T) {
	invalid := []MeasurementFlag{"", "unknown", "NORMAL"}
	for _, f := range invalid {
		if IsValidMeasurementFlag(f) {
			t.Errorf("expected %q to be invalid", f)
		}
	}
}

func TestNormalizeMarkerName_KnownAliases_ReturnsCanonical(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Glucose", "fasting_glucose"},
		{"Fasting Glucose", "fasting_glucose"},
		{"GLUCOSE, FASTING", "fasting_glucose"},
		{"Total Cholesterol", "total_cholesterol"},
		{"LDL Cholesterol", "ldl_cholesterol"},
		{"LDL-C", "ldl_cholesterol"},
		{"HDL Cholesterol", "hdl_cholesterol"},
		{"HDL-C", "hdl_cholesterol"},
		{"Triglycerides", "triglycerides"},
		{"HbA1c", "hba1c"},
		{"Hemoglobin A1C", "hba1c"},
		{"A1C", "hba1c"},
		{"Testosterone", "testosterone"},
		{"TSH", "tsh"},
		{"Vitamin D", "vitamin_d"},
		{"Iron", "iron"},
		{"  glucose  ", "fasting_glucose"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := NormalizeMarkerName(tc.input)
			if got != tc.expected {
				t.Errorf("NormalizeMarkerName(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestNormalizeMarkerName_UnknownName_ReturnsLowered(t *testing.T) {
	got := NormalizeMarkerName("Some Unknown Marker")
	if got != "some unknown marker" {
		t.Errorf("expected lowered name, got %q", got)
	}
}

func TestDetermineFlag_BothNil_ReturnsNil(t *testing.T) {
	flag := DetermineFlag(100, nil, nil)
	if flag != nil {
		t.Errorf("expected nil flag, got %v", *flag)
	}
}

func TestDetermineFlag_WithinRange_ReturnsNormal(t *testing.T) {
	low, high := 70.0, 100.0
	flag := DetermineFlag(85, &low, &high)
	if flag == nil || *flag != FlagNormal {
		t.Errorf("expected normal, got %v", flag)
	}
}

func TestDetermineFlag_BelowLow_ReturnsLow(t *testing.T) {
	low, high := 70.0, 100.0
	flag := DetermineFlag(50, &low, &high)
	if flag == nil || *flag != FlagLow {
		t.Errorf("expected low, got %v", flag)
	}
}

func TestDetermineFlag_AboveHigh_ReturnsHigh(t *testing.T) {
	low, high := 70.0, 100.0
	flag := DetermineFlag(142, &low, &high)
	if flag == nil || *flag != FlagHigh {
		t.Errorf("expected high, got %v", flag)
	}
}

func TestDetermineFlag_AtLowBound_ReturnsNormal(t *testing.T) {
	low, high := 70.0, 100.0
	flag := DetermineFlag(70, &low, &high)
	if flag == nil || *flag != FlagNormal {
		t.Errorf("expected normal at lower bound, got %v", flag)
	}
}

func TestDetermineFlag_AtHighBound_ReturnsNormal(t *testing.T) {
	low, high := 70.0, 100.0
	flag := DetermineFlag(100, &low, &high)
	if flag == nil || *flag != FlagNormal {
		t.Errorf("expected normal at upper bound, got %v", flag)
	}
}

func TestDetermineFlag_OnlyHighBound_BelowHigh_ReturnsNormal(t *testing.T) {
	high := 200.0
	flag := DetermineFlag(150, nil, &high)
	if flag == nil || *flag != FlagNormal {
		t.Errorf("expected normal, got %v", flag)
	}
}

func TestDetermineFlag_OnlyHighBound_AboveHigh_ReturnsHigh(t *testing.T) {
	high := 200.0
	flag := DetermineFlag(250, nil, &high)
	if flag == nil || *flag != FlagHigh {
		t.Errorf("expected high, got %v", flag)
	}
}

func TestDetermineFlag_OnlyLowBound_AboveLow_ReturnsNormal(t *testing.T) {
	low := 40.0
	flag := DetermineFlag(50, &low, nil)
	if flag == nil || *flag != FlagNormal {
		t.Errorf("expected normal, got %v", flag)
	}
}

func TestDetermineFlag_OnlyLowBound_BelowLow_ReturnsLow(t *testing.T) {
	low := 40.0
	flag := DetermineFlag(30, &low, nil)
	if flag == nil || *flag != FlagLow {
		t.Errorf("expected low, got %v", flag)
	}
}

func TestParseReferenceRange_Range_ReturnsLowAndHigh(t *testing.T) {
	low, high := ParseReferenceRange("70-100")
	if low == nil || *low != 70 {
		t.Errorf("expected low=70, got %v", low)
	}
	if high == nil || *high != 100 {
		t.Errorf("expected high=100, got %v", high)
	}
}

func TestParseReferenceRange_LessThan_ReturnsHighOnly(t *testing.T) {
	low, high := ParseReferenceRange("<200")
	if low != nil {
		t.Errorf("expected low=nil, got %v", *low)
	}
	if high == nil || *high != 200 {
		t.Errorf("expected high=200, got %v", high)
	}
}

func TestParseReferenceRange_GreaterThan_ReturnsLowOnly(t *testing.T) {
	low, high := ParseReferenceRange(">40")
	if low == nil || *low != 40 {
		t.Errorf("expected low=40, got %v", low)
	}
	if high != nil {
		t.Errorf("expected high=nil, got %v", *high)
	}
}

func TestParseReferenceRange_Decimal_ReturnsCorrectValues(t *testing.T) {
	low, high := ParseReferenceRange("0.4-4.0")
	if low == nil || *low != 0.4 {
		t.Errorf("expected low=0.4, got %v", low)
	}
	if high == nil || *high != 4.0 {
		t.Errorf("expected high=4.0, got %v", high)
	}
}

func TestParseReferenceRange_LessThanDecimal_ReturnsHighOnly(t *testing.T) {
	low, high := ParseReferenceRange("< 5.7")
	if low != nil {
		t.Errorf("expected low=nil, got %v", *low)
	}
	if high == nil || *high != 5.7 {
		t.Errorf("expected high=5.7, got %v", high)
	}
}

func TestParseReferenceRange_Empty_ReturnsNils(t *testing.T) {
	low, high := ParseReferenceRange("")
	if low != nil || high != nil {
		t.Errorf("expected both nil for empty string")
	}
}

func TestParseReferenceRange_Whitespace_ReturnsNils(t *testing.T) {
	low, high := ParseReferenceRange("   ")
	if low != nil || high != nil {
		t.Errorf("expected both nil for whitespace string")
	}
}

func TestParseReferenceRange_LargeRange_ReturnsCorrectValues(t *testing.T) {
	low, high := ParseReferenceRange("300-1000")
	if low == nil || *low != 300 {
		t.Errorf("expected low=300, got %v", low)
	}
	if high == nil || *high != 1000 {
		t.Errorf("expected high=1000, got %v", high)
	}
}

func TestParseReferenceRange_LessThanEqual_ReturnsHighOnly(t *testing.T) {
	low, high := ParseReferenceRange("<=150")
	if low != nil {
		t.Errorf("expected low=nil, got %v", *low)
	}
	if high == nil || *high != 150 {
		t.Errorf("expected high=150, got %v", high)
	}
}

func TestParseReferenceRange_GreaterThanEqual_ReturnsLowOnly(t *testing.T) {
	low, high := ParseReferenceRange(">=30")
	if low == nil || *low != 30 {
		t.Errorf("expected low=30, got %v", low)
	}
	if high != nil {
		t.Errorf("expected high=nil, got %v", *high)
	}
}

func TestParseFloat_ValidIntegers(t *testing.T) {
	v, ok := parseFloat("100")
	if !ok || v != 100 {
		t.Errorf("expected 100, got %v (ok=%v)", v, ok)
	}
}

func TestParseFloat_ValidDecimals(t *testing.T) {
	v, ok := parseFloat("3.14")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if v < 3.139 || v > 3.141 {
		t.Errorf("expected ~3.14, got %v", v)
	}
}

func TestParseFloat_Empty_ReturnsFalse(t *testing.T) {
	_, ok := parseFloat("")
	if ok {
		t.Error("expected ok=false for empty string")
	}
}

func TestParseFloat_NonNumeric_ReturnsFalse(t *testing.T) {
	_, ok := parseFloat("abc")
	if ok {
		t.Error("expected ok=false for non-numeric string")
	}
}

func TestParseFloat_Negative(t *testing.T) {
	v, ok := parseFloat("-5.5")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if v < -5.51 || v > -5.49 {
		t.Errorf("expected ~-5.5, got %v", v)
	}
}

func TestMeasurementFlagConstants_Values(t *testing.T) {
	tests := []struct {
		flag     MeasurementFlag
		expected string
	}{
		{FlagNormal, "normal"},
		{FlagLow, "low"},
		{FlagHigh, "high"},
		{FlagCriticalLow, "critical_low"},
		{FlagCriticalHigh, "critical_high"},
	}
	for _, tc := range tests {
		if string(tc.flag) != tc.expected {
			t.Errorf("expected %q, got %q", tc.expected, string(tc.flag))
		}
	}
}
