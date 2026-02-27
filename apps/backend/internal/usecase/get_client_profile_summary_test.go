package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestGetClientProfileSummary_Success_ReturnsConsolidatedProfile(t *testing.T) {
	now := time.Now()
	measurementRepo := &stubMeasurementRepo{
		findLatestResult: []*entities.Measurement{
			{ID: "m1", ClientID: "client-1", Type: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
			{ID: "m2", ClientID: "client-1", Type: "body_fat_pct", Value: 22.3, Unit: "%", MeasuredAt: now},
			{ID: "m3", ClientID: "client-1", Type: "ldl_cholesterol", Value: 142, Unit: "mg/dL", Flag: "high", MeasuredAt: now},
			{ID: "m4", ClientID: "client-1", Type: "calories", Value: 2150, Unit: "kcal", MeasuredAt: now},
			{ID: "m5", ClientID: "client-1", Type: "protein", Value: 165, Unit: "g", MeasuredAt: now},
		},
	}
	wearableRepo := &stubWearableRepo{
		findByClientResult: []*entities.WearableSummary{
			{
				ID: "w1", ClientID: "client-1", Source: "whoop", SummaryDate: "2026-02-27",
				Metrics: map[string]interface{}{"hrv": 45.0, "sleep_hours": 7.2, "recovery_score": 68.0},
			},
			{
				ID: "w2", ClientID: "client-1", Source: "whoop", SummaryDate: "2026-02-26",
				Metrics: map[string]interface{}{"hrv": 50.0, "sleep_hours": 8.0, "recovery_score": 72.0},
			},
		},
	}
	ccRepo := &stubCoachClientRepoForAuth{
		relationship: &entities.CoachClient{ID: "cc1", CoachID: "coach-1", ClientID: "client-1"},
	}
	auditRepo := &stubAuditRepoForMeasurements{}

	uc := NewGetClientProfileSummaryUseCase(measurementRepo, wearableRepo, ccRepo, auditRepo)
	result, err := uc.Execute(context.Background(), "coach-1", "client-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Body composition
	if weight, ok := result.BodyComposition["weight"]; !ok {
		t.Error("expected weight in body composition")
	} else if weight.Value != 185.4 {
		t.Errorf("expected weight 185.4, got %f", weight.Value)
	}

	if bf, ok := result.BodyComposition["body_fat_pct"]; !ok {
		t.Error("expected body_fat_pct in body composition")
	} else if bf.Value != 22.3 {
		t.Errorf("expected body_fat_pct 22.3, got %f", bf.Value)
	}

	// Labs
	if result.Labs.FlaggedCount != 1 {
		t.Errorf("expected 1 flagged lab, got %d", result.Labs.FlaggedCount)
	}
	if len(result.Labs.Markers) != 1 {
		t.Errorf("expected 1 lab marker, got %d", len(result.Labs.Markers))
	}

	// Wearable
	if result.Wearable.Source != "whoop" {
		t.Errorf("expected source whoop, got %s", result.Wearable.Source)
	}
	if result.Wearable.AvgHRV7d == nil {
		t.Error("expected non-nil AvgHRV7d")
	} else if *result.Wearable.AvgHRV7d != 47.5 {
		t.Errorf("expected avg HRV 47.5, got %f", *result.Wearable.AvgHRV7d)
	}

	// Nutrition
	if result.Nutrition.AvgCalories7d == nil {
		t.Error("expected non-nil AvgCalories7d")
	} else if *result.Nutrition.AvgCalories7d != 2150 {
		t.Errorf("expected avg calories 2150, got %f", *result.Nutrition.AvgCalories7d)
	}
	if result.Nutrition.AvgProtein7d == nil {
		t.Error("expected non-nil AvgProtein7d")
	} else if *result.Nutrition.AvgProtein7d != 165 {
		t.Errorf("expected avg protein 165, got %f", *result.Nutrition.AvgProtein7d)
	}

	// Audit
	if len(auditRepo.events) != 1 {
		t.Errorf("expected 1 audit event, got %d", len(auditRepo.events))
	}
}

func TestGetClientProfileSummary_Unauthorized_ReturnsAuthorizationError(t *testing.T) {
	ccRepo := &stubCoachClientRepoForAuth{relationship: nil}
	uc := NewGetClientProfileSummaryUseCase(
		&stubMeasurementRepo{},
		&stubWearableRepo{},
		ccRepo,
		&stubAuditRepoForMeasurements{},
	)

	_, err := uc.Execute(context.Background(), "coach-1", "client-1")

	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T: %v", err, err)
	}
}

func TestGetClientProfileSummary_MissingCoachID_ReturnsValidationError(t *testing.T) {
	uc := NewGetClientProfileSummaryUseCase(
		&stubMeasurementRepo{},
		&stubWearableRepo{},
		&stubCoachClientRepoForAuth{},
		&stubAuditRepoForMeasurements{},
	)

	_, err := uc.Execute(context.Background(), "", "client-1")

	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestGetClientProfileSummary_EmptyData_ReturnsEmptyProfile(t *testing.T) {
	measurementRepo := &stubMeasurementRepo{
		findLatestResult: []*entities.Measurement{},
	}
	wearableRepo := &stubWearableRepo{
		findByClientResult: []*entities.WearableSummary{},
	}
	ccRepo := &stubCoachClientRepoForAuth{
		relationship: &entities.CoachClient{ID: "cc1", CoachID: "coach-1", ClientID: "client-1"},
	}

	uc := NewGetClientProfileSummaryUseCase(measurementRepo, wearableRepo, ccRepo, &stubAuditRepoForMeasurements{})
	result, err := uc.Execute(context.Background(), "coach-1", "client-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.BodyComposition) != 0 {
		t.Errorf("expected empty body composition, got %d entries", len(result.BodyComposition))
	}
	if result.Labs.FlaggedCount != 0 {
		t.Errorf("expected 0 flagged labs, got %d", result.Labs.FlaggedCount)
	}
	if result.Wearable.Source != "" {
		t.Errorf("expected empty wearable source, got %s", result.Wearable.Source)
	}
}

func TestGetClientProfileSummary_MeasurementRepoError_ReturnsError(t *testing.T) {
	measurementRepo := &stubMeasurementRepo{
		findLatestErr: errors.New("db error"),
	}
	ccRepo := &stubCoachClientRepoForAuth{
		relationship: &entities.CoachClient{ID: "cc1", CoachID: "coach-1", ClientID: "client-1"},
	}
	uc := NewGetClientProfileSummaryUseCase(measurementRepo, &stubWearableRepo{}, ccRepo, &stubAuditRepoForMeasurements{})

	_, err := uc.Execute(context.Background(), "coach-1", "client-1")

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetClientProfileSummary_WearableRepoError_ReturnsError(t *testing.T) {
	measurementRepo := &stubMeasurementRepo{
		findLatestResult: []*entities.Measurement{},
	}
	wearableRepo := &stubWearableRepo{
		findByClientErr: errors.New("db error"),
	}
	ccRepo := &stubCoachClientRepoForAuth{
		relationship: &entities.CoachClient{ID: "cc1", CoachID: "coach-1", ClientID: "client-1"},
	}
	uc := NewGetClientProfileSummaryUseCase(measurementRepo, wearableRepo, ccRepo, &stubAuditRepoForMeasurements{})

	_, err := uc.Execute(context.Background(), "coach-1", "client-1")

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestBuildProfileSummary_ExtractFloat_IntTypes(t *testing.T) {
	wearables := []*entities.WearableSummary{
		{
			Source:      "garmin",
			SummaryDate: "2026-02-27",
			Metrics:     map[string]interface{}{"hrv": int(50), "sleep_hours": int64(8), "recovery_score": float32(75.0)},
		},
	}

	summary := buildProfileSummary([]*entities.Measurement{}, wearables)

	if summary.Wearable.AvgHRV7d == nil {
		t.Error("expected non-nil AvgHRV7d for int metrics")
	} else if *summary.Wearable.AvgHRV7d != 50 {
		t.Errorf("expected avg HRV 50, got %f", *summary.Wearable.AvgHRV7d)
	}
	if summary.Wearable.AvgSleep7d == nil {
		t.Error("expected non-nil AvgSleep7d for int64 metrics")
	} else if *summary.Wearable.AvgSleep7d != 8 {
		t.Errorf("expected avg sleep 8, got %f", *summary.Wearable.AvgSleep7d)
	}
	if summary.Wearable.AvgRecovery7d == nil {
		t.Error("expected non-nil AvgRecovery7d for float32 metrics")
	} else if *summary.Wearable.AvgRecovery7d != 75 {
		t.Errorf("expected avg recovery 75, got %f", *summary.Wearable.AvgRecovery7d)
	}
}

func TestBuildProfileSummary_ExtractFloat_InvalidType_Skipped(t *testing.T) {
	wearables := []*entities.WearableSummary{
		{
			Source:      "garmin",
			SummaryDate: "2026-02-27",
			Metrics:     map[string]interface{}{"hrv": "not-a-number"},
		},
	}

	summary := buildProfileSummary([]*entities.Measurement{}, wearables)

	if summary.Wearable.AvgHRV7d != nil {
		t.Error("expected nil AvgHRV7d for invalid metric type")
	}
}
