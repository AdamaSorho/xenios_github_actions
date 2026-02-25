package entities

import "testing"

func TestReferenceRange_EvaluateFlag_Normal_ReturnsNormal(t *testing.T) {
	ref := &ReferenceRange{
		LowCritical:  40,
		LowNormal:    70,
		HighNormal:   100,
		HighCritical: 200,
	}

	flag := ref.EvaluateFlag(85)
	if flag != MeasurementFlagNormal {
		t.Errorf("expected normal, got %q", flag)
	}
}

func TestReferenceRange_EvaluateFlag_High_ReturnsHigh(t *testing.T) {
	ref := &ReferenceRange{
		LowCritical:  40,
		LowNormal:    70,
		HighNormal:   100,
		HighCritical: 200,
	}

	flag := ref.EvaluateFlag(150)
	if flag != MeasurementFlagHigh {
		t.Errorf("expected high, got %q", flag)
	}
}

func TestReferenceRange_EvaluateFlag_CriticalHigh_ReturnsCriticalHigh(t *testing.T) {
	ref := &ReferenceRange{
		LowCritical:  40,
		LowNormal:    70,
		HighNormal:   100,
		HighCritical: 200,
	}

	flag := ref.EvaluateFlag(250)
	if flag != MeasurementFlagCriticalHigh {
		t.Errorf("expected critical_high, got %q", flag)
	}
}

func TestReferenceRange_EvaluateFlag_Low_ReturnsLow(t *testing.T) {
	ref := &ReferenceRange{
		LowCritical:  40,
		LowNormal:    70,
		HighNormal:   100,
		HighCritical: 200,
	}

	flag := ref.EvaluateFlag(50)
	if flag != MeasurementFlagLow {
		t.Errorf("expected low, got %q", flag)
	}
}

func TestReferenceRange_EvaluateFlag_CriticalLow_ReturnsCriticalLow(t *testing.T) {
	ref := &ReferenceRange{
		LowCritical:  40,
		LowNormal:    70,
		HighNormal:   100,
		HighCritical: 200,
	}

	flag := ref.EvaluateFlag(30)
	if flag != MeasurementFlagCriticalLow {
		t.Errorf("expected critical_low, got %q", flag)
	}
}

func TestReferenceRange_EvaluateFlag_BoundaryHigh_ReturnsNormal(t *testing.T) {
	ref := &ReferenceRange{
		LowCritical:  40,
		LowNormal:    70,
		HighNormal:   100,
		HighCritical: 200,
	}

	// Value exactly at high normal boundary should be normal
	flag := ref.EvaluateFlag(100)
	if flag != MeasurementFlagNormal {
		t.Errorf("expected normal at boundary, got %q", flag)
	}
}

func TestReferenceRange_EvaluateFlag_BoundaryCritical_ReturnsCriticalLow(t *testing.T) {
	ref := &ReferenceRange{
		LowCritical:  40,
		LowNormal:    70,
		HighNormal:   100,
		HighCritical: 200,
	}

	// Value exactly at critical low boundary should be critical low
	flag := ref.EvaluateFlag(40)
	if flag != MeasurementFlagCriticalLow {
		t.Errorf("expected critical_low at boundary, got %q", flag)
	}
}

func TestLabReferenceRanges_ContainsExpectedMarkers(t *testing.T) {
	expected := []string{
		"ldl_cholesterol",
		"hdl_cholesterol",
		"fasting_glucose",
		"total_cholesterol",
		"triglycerides",
		"hemoglobin_a1c",
	}

	for _, marker := range expected {
		if _, ok := LabReferenceRanges[marker]; !ok {
			t.Errorf("expected LabReferenceRanges to contain %q", marker)
		}
	}
}

func TestLabReferenceRanges_LDLHighValue_ReturnsHigh(t *testing.T) {
	ref := LabReferenceRanges["ldl_cholesterol"]
	flag := ref.EvaluateFlag(142)
	if flag != MeasurementFlagHigh {
		t.Errorf("expected high for LDL 142, got %q", flag)
	}
}

func TestLabReferenceRanges_GlucoseCriticalHigh_ReturnsCriticalHigh(t *testing.T) {
	ref := LabReferenceRanges["fasting_glucose"]
	flag := ref.EvaluateFlag(250)
	if flag != MeasurementFlagCriticalHigh {
		t.Errorf("expected critical_high for glucose 250, got %q", flag)
	}
}
