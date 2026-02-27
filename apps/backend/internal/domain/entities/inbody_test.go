package entities

import "testing"

func TestParseInBodyText_FullReport_ExtractsAllMetrics(t *testing.T) {
	text := `InBody 570 Body Composition Analysis

Weight: 185.4 lbs
Skeletal Muscle Mass: 78.2 lbs
Body Fat Percentage: 22.3 %
Basal Metabolic Rate: 1847 kcal
Total Body Water: 42.1 L
Lean Body Mass: 144.1 lbs`

	result, err := ParseInBodyText(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Partial {
		t.Error("expected full extraction, got partial")
	}
	if len(result.Metrics) != 6 {
		t.Fatalf("expected 6 metrics, got %d", len(result.Metrics))
	}

	expected := map[MeasurementType]struct{ value float64; unit string }{
		MeasurementTypeWeight:             {185.4, "lbs"},
		MeasurementTypeSkeletalMuscleMass: {78.2, "lbs"},
		MeasurementTypeBodyFatPct:         {22.3, "%"},
		MeasurementTypeBMR:                {1847, "kcal"},
		MeasurementTypeTotalBodyWater:     {42.1, "L"},
		MeasurementTypeLeanBodyMass:       {144.1, "lbs"},
	}

	for _, m := range result.Metrics {
		exp, ok := expected[m.Type]
		if !ok {
			t.Errorf("unexpected metric type: %s", m.Type)
			continue
		}
		if m.Value != exp.value {
			t.Errorf("metric %s: expected value %v, got %v", m.Type, exp.value, m.Value)
		}
		if m.Unit != exp.unit {
			t.Errorf("metric %s: expected unit %q, got %q", m.Type, exp.unit, m.Unit)
		}
	}
}

func TestParseInBodyText_KgUnits_ExtractsCorrectly(t *testing.T) {
	text := `Body Composition Analysis
Weight: 84.1 kg
Skeletal Muscle Mass: 35.5 kg
Body Fat: 18.7 %
BMR: 1750 kcal
Total Body Water: 45.2 L
Lean Body Mass: 68.3 kg`

	result, err := ParseInBodyText(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Partial {
		t.Error("expected full extraction, got partial")
	}

	found := map[MeasurementType]bool{}
	for _, m := range result.Metrics {
		found[m.Type] = true
		if m.Type == MeasurementTypeWeight && (m.Value != 84.1 || m.Unit != "kg") {
			t.Errorf("weight: expected 84.1 kg, got %v %s", m.Value, m.Unit)
		}
	}
	if !found[MeasurementTypeWeight] {
		t.Error("expected weight metric")
	}
}

func TestParseInBodyText_PartialReport_FlagsPartial(t *testing.T) {
	text := `Weight: 185.4 lbs
Body Fat: 22.3 %`

	result, err := ParseInBodyText(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Partial {
		t.Error("expected partial flag to be set")
	}
	if len(result.Metrics) != 2 {
		t.Errorf("expected 2 metrics, got %d", len(result.Metrics))
	}
}

func TestParseInBodyText_EmptyText_ReturnsError(t *testing.T) {
	_, err := ParseInBodyText("")
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestParseInBodyText_NoMetricsFound_ReturnsError(t *testing.T) {
	_, err := ParseInBodyText("This is not an InBody report at all.")
	if err == nil {
		t.Error("expected error when no metrics found")
	}
}

func TestParseInBodyText_AlternativeFormats_ExtractsCorrectly(t *testing.T) {
	text := `InBody770 Result Sheet
Weight        84.1 kg
SMM           35.5 kg
PBF           18.7 %
BMR           1750 kcal
TBW           45.2 L
LBM           68.3 kg`

	result, err := ParseInBodyText(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := map[MeasurementType]bool{}
	for _, m := range result.Metrics {
		found[m.Type] = true
	}
	if !found[MeasurementTypeWeight] {
		t.Error("expected weight metric from 'Weight' pattern")
	}
	if !found[MeasurementTypeSkeletalMuscleMass] {
		t.Error("expected SMM metric from 'SMM' pattern")
	}
	if !found[MeasurementTypeBodyFatPct] {
		t.Error("expected body fat metric from 'PBF' pattern")
	}
}

func TestParseInBodyText_TabSeparated_ExtractsCorrectly(t *testing.T) {
	text := "Weight\t85.0\tkg\nBody Fat %\t20.5\t%\nSkeletal Muscle Mass\t36.0\tkg"

	result, err := ParseInBodyText(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Metrics) < 2 {
		t.Errorf("expected at least 2 metrics, got %d", len(result.Metrics))
	}
}

func TestParseInBodyText_DecimalVariations_ParsesCorrectly(t *testing.T) {
	text := `Weight: 85 kg
Body Fat: 20.5 %
BMR: 1847 kcal`

	result, err := ParseInBodyText(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, m := range result.Metrics {
		if m.Type == MeasurementTypeWeight && m.Value != 85 {
			t.Errorf("expected weight 85, got %v", m.Value)
		}
		if m.Type == MeasurementTypeBMR && m.Value != 1847 {
			t.Errorf("expected BMR 1847, got %v", m.Value)
		}
	}
}

func TestParseInBodyText_DuplicateMetrics_KeepsFirst(t *testing.T) {
	text := `Weight: 85.0 kg
Weight: 90.0 kg
Body Fat: 20.5 %`

	result, err := ParseInBodyText(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, m := range result.Metrics {
		if m.Type == MeasurementTypeWeight && m.Value != 85.0 {
			t.Errorf("expected first weight value 85.0, got %v", m.Value)
		}
	}
}
