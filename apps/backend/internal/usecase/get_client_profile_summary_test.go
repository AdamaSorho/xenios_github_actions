package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestGetClientProfileSummary_Success(t *testing.T) {
	now := time.Now()
	measurements := []*entities.Measurement{
		{ID: "m1", Type: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
		{ID: "m2", Type: "body_fat_pct", Value: 22.3, Unit: "%", MeasuredAt: now},
		{ID: "m3", Type: "ldl_cholesterol", Value: 142, Unit: "mg/dL", MeasuredAt: now, Flag: strPtr("high")},
	}
	wearableSummaries := []*entities.WearableSummary{
		{ID: "w1", Source: "whoop", SummaryDate: "2026-01-15", Metrics: map[string]interface{}{"hrv": 45.0, "sleep_hours": 7.2, "recovery_score": 68.0}},
		{ID: "w2", Source: "whoop", SummaryDate: "2026-01-14", Metrics: map[string]interface{}{"hrv": 50.0, "sleep_hours": 7.0, "recovery_score": 72.0}},
	}

	measurementRepo := &mockMeasurementRepo{
		findLatestByClientIDFunc: func(ctx context.Context, clientID string) ([]*entities.Measurement, error) {
			return measurements, nil
		},
	}
	wearableRepo := &mockWearableRepo{
		findByClientIDFunc: func(ctx context.Context, clientID string, days int) ([]*entities.WearableSummary, error) {
			return wearableSummaries, nil
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetClientProfileSummaryUseCase(measurementRepo, wearableRepo, ccRepo, auditRepo)
	out, err := uc.Execute(context.Background(), GetClientProfileSummaryInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check body composition
	if w, ok := out.BodyComposition["weight"]; !ok {
		t.Error("expected weight in body composition")
	} else if w.Value != 185.4 {
		t.Errorf("expected weight 185.4, got %f", w.Value)
	}

	if bf, ok := out.BodyComposition["body_fat_pct"]; !ok {
		t.Error("expected body_fat_pct in body composition")
	} else if bf.Value != 22.3 {
		t.Errorf("expected body_fat_pct 22.3, got %f", bf.Value)
	}

	// Check labs
	if out.Labs.FlaggedCount != 1 {
		t.Errorf("expected 1 flagged lab, got %d", out.Labs.FlaggedCount)
	}
	if len(out.Labs.Markers) != 1 {
		t.Errorf("expected 1 lab marker, got %d", len(out.Labs.Markers))
	}
	if len(out.Labs.Markers) > 0 && out.Labs.Markers[0].Flag != "high" {
		t.Errorf("expected flag 'high', got '%s'", out.Labs.Markers[0].Flag)
	}

	// Check wearable
	if out.Wearable == nil {
		t.Fatal("expected non-nil wearable summary")
	}
	if out.Wearable.Source != "whoop" {
		t.Errorf("expected source 'whoop', got '%s'", out.Wearable.Source)
	}
	expectedHrv := 47.5 // (45 + 50) / 2
	if out.Wearable.AvgHrv7d != expectedHrv {
		t.Errorf("expected avg HRV %f, got %f", expectedHrv, out.Wearable.AvgHrv7d)
	}
	expectedSleep := 7.1 // (7.2 + 7.0) / 2
	if out.Wearable.AvgSleep7d != expectedSleep {
		t.Errorf("expected avg sleep %f, got %f", expectedSleep, out.Wearable.AvgSleep7d)
	}
	expectedRecovery := 70.0 // (68 + 72) / 2
	if out.Wearable.AvgRecovery7d != expectedRecovery {
		t.Errorf("expected avg recovery %f, got %f", expectedRecovery, out.Wearable.AvgRecovery7d)
	}
}

func TestGetClientProfileSummary_EmptyData(t *testing.T) {
	measurementRepo := &mockMeasurementRepo{
		findLatestByClientIDFunc: func(ctx context.Context, clientID string) ([]*entities.Measurement, error) {
			return nil, nil
		},
	}
	wearableRepo := &mockWearableRepo{
		findByClientIDFunc: func(ctx context.Context, clientID string, days int) ([]*entities.WearableSummary, error) {
			return nil, nil
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetClientProfileSummaryUseCase(measurementRepo, wearableRepo, ccRepo, auditRepo)
	out, err := uc.Execute(context.Background(), GetClientProfileSummaryInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.BodyComposition) != 0 {
		t.Errorf("expected empty body composition, got %d entries", len(out.BodyComposition))
	}
	if len(out.Labs.Markers) != 0 {
		t.Errorf("expected empty lab markers, got %d", len(out.Labs.Markers))
	}
	if out.Wearable != nil {
		t.Error("expected nil wearable summary with no data")
	}
}

func TestGetClientProfileSummary_Unauthorized(t *testing.T) {
	measurementRepo := &mockMeasurementRepo{}
	wearableRepo := &mockWearableRepo{}
	ccRepo := &mockCoachClientRepo{
		findByCoachAndClient: func(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
			return nil, nil
		},
	}
	auditRepo := &mockAuditRepo{}

	uc := NewGetClientProfileSummaryUseCase(measurementRepo, wearableRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetClientProfileSummaryInput{
		CoachID:  "coach-1",
		ClientID: "client-2",
	})

	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T", err)
	}
}

func TestGetClientProfileSummary_EmptyCoachID(t *testing.T) {
	uc := NewGetClientProfileSummaryUseCase(&mockMeasurementRepo{}, &mockWearableRepo{}, &mockCoachClientRepo{}, &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), GetClientProfileSummaryInput{
		ClientID: "client-1",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetClientProfileSummary_EmptyClientID(t *testing.T) {
	uc := NewGetClientProfileSummaryUseCase(&mockMeasurementRepo{}, &mockWearableRepo{}, &mockCoachClientRepo{}, &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), GetClientProfileSummaryInput{
		CoachID: "coach-1",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetClientProfileSummary_MeasurementRepoError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	measurementRepo := &mockMeasurementRepo{
		findLatestByClientIDFunc: func(ctx context.Context, clientID string) ([]*entities.Measurement, error) {
			return nil, repoErr
		},
	}
	wearableRepo := &mockWearableRepo{}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetClientProfileSummaryUseCase(measurementRepo, wearableRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetClientProfileSummaryInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if !errors.Is(err, repoErr) {
		t.Errorf("expected repository error, got %v", err)
	}
}

func TestGetClientProfileSummary_WearableRepoError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	measurementRepo := &mockMeasurementRepo{
		findLatestByClientIDFunc: func(ctx context.Context, clientID string) ([]*entities.Measurement, error) {
			return []*entities.Measurement{}, nil
		},
	}
	wearableRepo := &mockWearableRepo{
		findByClientIDFunc: func(ctx context.Context, clientID string, days int) ([]*entities.WearableSummary, error) {
			return nil, repoErr
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetClientProfileSummaryUseCase(measurementRepo, wearableRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetClientProfileSummaryInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if !errors.Is(err, repoErr) {
		t.Errorf("expected repository error, got %v", err)
	}
}

func TestGetClientProfileSummary_AuditLogged(t *testing.T) {
	measurementRepo := &mockMeasurementRepo{
		findLatestByClientIDFunc: func(ctx context.Context, clientID string) ([]*entities.Measurement, error) {
			return []*entities.Measurement{}, nil
		},
	}
	wearableRepo := &mockWearableRepo{}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetClientProfileSummaryUseCase(measurementRepo, wearableRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetClientProfileSummaryInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(auditRepo.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditRepo.events))
	}
	if auditRepo.events[0].Action != "phi.profile_summary_accessed" {
		t.Errorf("expected action 'phi.profile_summary_accessed', got '%s'", auditRepo.events[0].Action)
	}
}

func TestBuildProfileSummary_NoWearableData(t *testing.T) {
	output := buildProfileSummary(nil, nil)
	if output.Wearable != nil {
		t.Error("expected nil wearable with no data")
	}
	if len(output.BodyComposition) != 0 {
		t.Error("expected empty body composition")
	}
}

func TestGetMetricFloat_Float64(t *testing.T) {
	metrics := map[string]interface{}{"hrv": 45.0}
	v, ok := getMetricFloat(metrics, "hrv")
	if !ok || v != 45.0 {
		t.Errorf("expected 45.0, got %f ok=%v", v, ok)
	}
}

func TestGetMetricFloat_Int(t *testing.T) {
	metrics := map[string]interface{}{"hrv": 45}
	v, ok := getMetricFloat(metrics, "hrv")
	if !ok || v != 45.0 {
		t.Errorf("expected 45.0, got %f ok=%v", v, ok)
	}
}

func TestGetMetricFloat_Missing(t *testing.T) {
	metrics := map[string]interface{}{}
	_, ok := getMetricFloat(metrics, "hrv")
	if ok {
		t.Error("expected ok=false for missing key")
	}
}

func TestGetMetricFloat_Invalid(t *testing.T) {
	metrics := map[string]interface{}{"hrv": "not a number"}
	_, ok := getMetricFloat(metrics, "hrv")
	if ok {
		t.Error("expected ok=false for invalid type")
	}
}

func TestGetMetricFloat_Int64(t *testing.T) {
	metrics := map[string]interface{}{"hrv": int64(45)}
	v, ok := getMetricFloat(metrics, "hrv")
	if !ok || v != 45.0 {
		t.Errorf("expected 45.0, got %f ok=%v", v, ok)
	}
}

func TestGetClientProfileSummary_CoachClientRepoError(t *testing.T) {
	repoErr := errors.New("coach client lookup failed")
	measurementRepo := &mockMeasurementRepo{}
	wearableRepo := &mockWearableRepo{}
	ccRepo := &mockCoachClientRepo{
		findByCoachAndClient: func(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
			return nil, repoErr
		},
	}
	auditRepo := &mockAuditRepo{}

	uc := NewGetClientProfileSummaryUseCase(measurementRepo, wearableRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetClientProfileSummaryInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if !errors.Is(err, repoErr) {
		t.Errorf("expected coach client repo error to propagate, got %v", err)
	}
}

func TestBuildProfileSummary_NutritionTypes(t *testing.T) {
	now := time.Now()
	measurements := []*entities.Measurement{
		{Type: "calories", Value: 2100, Unit: "kcal", MeasuredAt: now},
		{Type: "protein", Value: 165, Unit: "g", MeasuredAt: now},
	}
	output := buildProfileSummary(measurements, nil)
	if output.Nutrition.AvgCalories7d != 2100 {
		t.Errorf("expected avg calories 2100, got %f", output.Nutrition.AvgCalories7d)
	}
	if output.Nutrition.AvgProtein7d != 165 {
		t.Errorf("expected avg protein 165, got %f", output.Nutrition.AvgProtein7d)
	}
}

func TestBuildProfileSummary_LabNoFlag(t *testing.T) {
	now := time.Now()
	measurements := []*entities.Measurement{
		{Type: "ldl_cholesterol", Value: 100, Unit: "mg/dL", MeasuredAt: now},
	}
	output := buildProfileSummary(measurements, nil)
	if output.Labs.FlaggedCount != 0 {
		t.Errorf("expected 0 flagged, got %d", output.Labs.FlaggedCount)
	}
	if len(output.Labs.Markers) != 1 {
		t.Fatalf("expected 1 marker, got %d", len(output.Labs.Markers))
	}
	if output.Labs.Markers[0].Flag != "" {
		t.Errorf("expected empty flag, got %q", output.Labs.Markers[0].Flag)
	}
}

func TestBuildProfileSummary_WearableMissingMetrics(t *testing.T) {
	summaries := []*entities.WearableSummary{
		{Source: "whoop", Metrics: map[string]interface{}{}},
	}
	output := buildProfileSummary(nil, summaries)
	if output.Wearable == nil {
		t.Fatal("expected non-nil wearable")
	}
	if output.Wearable.AvgHrv7d != 0 {
		t.Errorf("expected avg HRV 0, got %f", output.Wearable.AvgHrv7d)
	}
	if output.Wearable.AvgSleep7d != 0 {
		t.Errorf("expected avg sleep 0, got %f", output.Wearable.AvgSleep7d)
	}
	if output.Wearable.AvgRecovery7d != 0 {
		t.Errorf("expected avg recovery 0, got %f", output.Wearable.AvgRecovery7d)
	}
}

func strPtr(s string) *string {
	return &s
}
