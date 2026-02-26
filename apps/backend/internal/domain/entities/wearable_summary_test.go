package entities

import "testing"

func TestProfileSummary_EmptyFields_NoNilPanic(t *testing.T) {
	ps := &ProfileSummary{
		BodyComposition: map[string]*LatestMeasurement{},
		Labs:            &LabSummary{Markers: []*LatestMeasurement{}},
		Wearable:        &WearableAverages{},
		Nutrition:       &NutritionAverages{},
	}
	if ps.BodyComposition == nil {
		t.Error("expected non-nil body composition map")
	}
	if ps.Labs.FlaggedCount != 0 {
		t.Errorf("expected 0 flagged, got %d", ps.Labs.FlaggedCount)
	}
	if len(ps.Labs.Markers) != 0 {
		t.Errorf("expected 0 markers, got %d", len(ps.Labs.Markers))
	}
}

func TestWearableSummary_Fields(t *testing.T) {
	ws := &WearableSummary{
		ID:       "ws-1",
		ClientID: "c-1",
		Source:   "whoop",
		Metrics:  map[string]interface{}{"hrv": 45.2},
	}
	if ws.Source != "whoop" {
		t.Errorf("expected source whoop, got %s", ws.Source)
	}
	if ws.Metrics["hrv"] != 45.2 {
		t.Errorf("expected hrv 45.2, got %v", ws.Metrics["hrv"])
	}
}
