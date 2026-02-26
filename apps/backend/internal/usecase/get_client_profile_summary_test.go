package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestGetClientProfileSummary_Success_ReturnsConsolidated(t *testing.T) {
	now := time.Now()
	measRepo := &mockMeasurementRepo{
		findLatestFunc: func(ctx context.Context, clientID string) ([]*entities.Measurement, error) {
			return []*entities.Measurement{
				{ID: "m1", MeasurementType: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
				{ID: "m2", MeasurementType: "body_fat_pct", Value: 22.3, Unit: "%", MeasuredAt: now},
				{ID: "m3", MeasurementType: "ldl_cholesterol", Value: 142, Unit: "mg/dL", Flag: "high", MeasuredAt: now},
				{ID: "m4", MeasurementType: "calories", Value: 2150, Unit: "kcal", MeasuredAt: now},
				{ID: "m5", MeasurementType: "protein", Value: 165, Unit: "g", MeasuredAt: now},
			}, nil
		},
	}
	wearableRepo := &mockWearableRepo{
		findByClientFunc: func(ctx context.Context, clientID string, limit, offset int) ([]*entities.WearableSummary, error) {
			return []*entities.WearableSummary{
				{Source: "whoop", Metrics: map[string]interface{}{"hrv": 45.0, "sleep_hours": 7.2, "recovery_score": 68.0}},
				{Source: "whoop", Metrics: map[string]interface{}{"hrv": 50.0, "sleep_hours": 7.0, "recovery_score": 72.0}},
			}, nil
		},
	}
	auditRepo := &mockAuditRepo{}
	uc := NewGetClientProfileSummaryUseCase(measRepo, wearableRepo, authorizedCCRepo(), auditRepo)

	out, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.BodyComposition.Weight == nil {
		t.Fatal("expected weight in body composition")
	}
	if out.BodyComposition.Weight.Value != 185.4 {
		t.Errorf("expected weight 185.4, got %f", out.BodyComposition.Weight.Value)
	}

	if out.BodyComposition.BodyFatPct == nil {
		t.Fatal("expected body_fat_pct in body composition")
	}
	if out.BodyComposition.BodyFatPct.Value != 22.3 {
		t.Errorf("expected body fat 22.3, got %f", out.BodyComposition.BodyFatPct.Value)
	}

	if out.Labs.FlaggedCount != 1 {
		t.Errorf("expected 1 flagged lab, got %d", out.Labs.FlaggedCount)
	}
	if len(out.Labs.Markers) != 1 {
		t.Errorf("expected 1 marker, got %d", len(out.Labs.Markers))
	}

	if out.Wearable.Source != "whoop" {
		t.Errorf("expected source whoop, got %s", out.Wearable.Source)
	}
	expectedHRV := 47.5
	if out.Wearable.AvgHRV7d != expectedHRV {
		t.Errorf("expected avg HRV %f, got %f", expectedHRV, out.Wearable.AvgHRV7d)
	}

	if out.Nutrition.AvgCalories7d != 2150 {
		t.Errorf("expected avg calories 2150, got %f", out.Nutrition.AvgCalories7d)
	}
	if out.Nutrition.AvgProtein7d != 165 {
		t.Errorf("expected avg protein 165, got %f", out.Nutrition.AvgProtein7d)
	}
}

func TestGetClientProfileSummary_Unauthorized_ReturnsForbidden(t *testing.T) {
	uc := NewGetClientProfileSummaryUseCase(&mockMeasurementRepo{}, &mockWearableRepo{}, unauthorizedCCRepo(), &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err == nil {
		t.Fatal("expected error for unauthorized access")
	}
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func TestGetClientProfileSummary_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	uc := NewGetClientProfileSummaryUseCase(&mockMeasurementRepo{}, &mockWearableRepo{}, authorizedCCRepo(), &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), "", "client-1")
	if err == nil {
		t.Fatal("expected error for empty coach_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetClientProfileSummary_EmptyClientID_ReturnsValidationError(t *testing.T) {
	uc := NewGetClientProfileSummaryUseCase(&mockMeasurementRepo{}, &mockWearableRepo{}, authorizedCCRepo(), &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), "coach-1", "")
	if err == nil {
		t.Fatal("expected error for empty client_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetClientProfileSummary_EmptyData_ReturnsEmptyFields(t *testing.T) {
	measRepo := &mockMeasurementRepo{
		findLatestFunc: func(ctx context.Context, clientID string) ([]*entities.Measurement, error) {
			return []*entities.Measurement{}, nil
		},
	}
	wearableRepo := &mockWearableRepo{
		findByClientFunc: func(ctx context.Context, clientID string, limit, offset int) ([]*entities.WearableSummary, error) {
			return []*entities.WearableSummary{}, nil
		},
	}
	uc := NewGetClientProfileSummaryUseCase(measRepo, wearableRepo, authorizedCCRepo(), &mockAuditRepo{})
	out, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.BodyComposition.Weight != nil {
		t.Error("expected nil weight for empty data")
	}
	if out.Labs.FlaggedCount != 0 {
		t.Errorf("expected 0 flagged labs, got %d", out.Labs.FlaggedCount)
	}
	if out.Wearable.Source != "" {
		t.Errorf("expected empty wearable source, got %s", out.Wearable.Source)
	}
}

func TestGetClientProfileSummary_MeasurementRepoError_Propagates(t *testing.T) {
	repoErr := errors.New("database unavailable")
	measRepo := &mockMeasurementRepo{
		findLatestFunc: func(ctx context.Context, clientID string) ([]*entities.Measurement, error) {
			return nil, repoErr
		},
	}
	uc := NewGetClientProfileSummaryUseCase(measRepo, &mockWearableRepo{}, authorizedCCRepo(), &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err == nil {
		t.Fatal("expected error from repository")
	}
}

func TestGetClientProfileSummary_WearableRepoError_Propagates(t *testing.T) {
	repoErr := errors.New("database unavailable")
	wearableRepo := &mockWearableRepo{
		findByClientFunc: func(ctx context.Context, clientID string, limit, offset int) ([]*entities.WearableSummary, error) {
			return nil, repoErr
		},
	}
	uc := NewGetClientProfileSummaryUseCase(&mockMeasurementRepo{}, wearableRepo, authorizedCCRepo(), &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err == nil {
		t.Fatal("expected error from repository")
	}
}

func TestGetClientProfileSummary_AuditEvent_Logged(t *testing.T) {
	auditRepo := &mockAuditRepo{}
	uc := NewGetClientProfileSummaryUseCase(&mockMeasurementRepo{}, &mockWearableRepo{}, authorizedCCRepo(), auditRepo)
	_, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(auditRepo.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditRepo.events))
	}
	if auditRepo.events[0].Action != "phi.access" {
		t.Errorf("expected phi.access action, got %s", auditRepo.events[0].Action)
	}
}

func TestBuildBodyComposition_SkeletalMuscleMass(t *testing.T) {
	now := time.Now()
	measurements := []*entities.Measurement{
		{MeasurementType: "skeletal_muscle_mass", Value: 78.2, Unit: "lbs", MeasuredAt: now},
	}
	bc := buildBodyComposition(measurements)
	if bc.SkeletalMuscleMass == nil {
		t.Fatal("expected skeletal muscle mass")
	}
	if bc.SkeletalMuscleMass.Value != 78.2 {
		t.Errorf("expected 78.2, got %f", bc.SkeletalMuscleMass.Value)
	}
}

func TestBuildWearableSummary_Empty(t *testing.T) {
	ws := buildWearableSummary([]*entities.WearableSummary{})
	if ws.Source != "" {
		t.Errorf("expected empty source, got %s", ws.Source)
	}
	if ws.AvgHRV7d != 0 {
		t.Errorf("expected 0 HRV, got %f", ws.AvgHRV7d)
	}
}

func TestExtractFloat_IntValue(t *testing.T) {
	metrics := map[string]interface{}{"hrv": 45}
	v, ok := extractFloat(metrics, "hrv")
	if !ok {
		t.Fatal("expected ok")
	}
	if v != 45 {
		t.Errorf("expected 45, got %f", v)
	}
}

func TestExtractFloat_MissingKey(t *testing.T) {
	metrics := map[string]interface{}{"hrv": 45.0}
	_, ok := extractFloat(metrics, "nonexistent")
	if ok {
		t.Error("expected not ok for missing key")
	}
}

func TestExtractFloat_UnsupportedType(t *testing.T) {
	metrics := map[string]interface{}{"hrv": "not a number"}
	_, ok := extractFloat(metrics, "hrv")
	if ok {
		t.Error("expected not ok for string type")
	}
}
