package entities

import "testing"

func TestWearableSummary_GetMetricFloat64_ExistingFloat_ReturnsValue(t *testing.T) {
	ws := &WearableSummary{
		Metrics: map[string]interface{}{"hrv": float64(55.5)},
	}
	val, ok := ws.GetMetricFloat64("hrv")
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if val != 55.5 {
		t.Errorf("expected 55.5, got %f", val)
	}
}

func TestWearableSummary_GetMetricFloat64_ExistingInt_ReturnsValue(t *testing.T) {
	ws := &WearableSummary{
		Metrics: map[string]interface{}{"steps": int(10000)},
	}
	val, ok := ws.GetMetricFloat64("steps")
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if val != 10000 {
		t.Errorf("expected 10000, got %f", val)
	}
}

func TestWearableSummary_GetMetricFloat64_ExistingInt64_ReturnsValue(t *testing.T) {
	ws := &WearableSummary{
		Metrics: map[string]interface{}{"calories": int64(2500)},
	}
	val, ok := ws.GetMetricFloat64("calories")
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if val != 2500 {
		t.Errorf("expected 2500, got %f", val)
	}
}

func TestWearableSummary_GetMetricFloat64_Missing_ReturnsFalse(t *testing.T) {
	ws := &WearableSummary{
		Metrics: map[string]interface{}{"hrv": float64(55)},
	}
	_, ok := ws.GetMetricFloat64("nonexistent")
	if ok {
		t.Error("expected ok to be false for missing key")
	}
}

func TestWearableSummary_GetMetricFloat64_StringValue_ReturnsFalse(t *testing.T) {
	ws := &WearableSummary{
		Metrics: map[string]interface{}{"source": "whoop"},
	}
	_, ok := ws.GetMetricFloat64("source")
	if ok {
		t.Error("expected ok to be false for string value")
	}
}

func TestWearableSummary_GetMetricFloat64_NilMetrics_ReturnsFalse(t *testing.T) {
	ws := &WearableSummary{
		Metrics: nil,
	}
	_, ok := ws.GetMetricFloat64("hrv")
	if ok {
		t.Error("expected ok to be false for nil metrics")
	}
}
