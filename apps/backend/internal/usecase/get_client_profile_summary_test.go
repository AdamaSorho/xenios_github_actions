package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func newProfileSummaryUseCase(mRepo *mockMeasurementRepo, wRepo *mockWearableSummaryRepo, ccRepo *mockCoachClientRepo, aRepo *mockAuditRepo) *GetClientProfileSummaryUseCase {
	return NewGetClientProfileSummaryUseCase(mRepo, wRepo, ccRepo, aRepo)
}

func TestGetClientProfileSummary_Success(t *testing.T) {
	now := time.Now()
	mRepo := &mockMeasurementRepo{
		findLatestByClientIDFunc: func(ctx context.Context, clientID string) ([]*entities.LatestMeasurement, error) {
			return []*entities.LatestMeasurement{
				{MeasurementType: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
				{MeasurementType: "body_fat_pct", Value: 22.3, Unit: "%", MeasuredAt: now},
				{MeasurementType: "ldl_cholesterol", Value: 142, Unit: "mg/dL", MeasuredAt: now},
				{MeasurementType: "calories", Value: 2150, Unit: "kcal", MeasuredAt: now},
				{MeasurementType: "protein", Value: 165, Unit: "g", MeasuredAt: now},
			}, nil
		},
	}
	hrv := 45.2
	sleep := 7.2
	recovery := 68.0
	wRepo := &mockWearableSummaryRepo{
		findAveragesFunc: func(ctx context.Context, clientID string, days int) (*entities.WearableAverages, error) {
			return &entities.WearableAverages{
				Source:        "whoop",
				AvgHRV7d:      &hrv,
				AvgSleep7d:    &sleep,
				AvgRecovery7d: &recovery,
			}, nil
		},
	}
	uc := newProfileSummaryUseCase(mRepo, wRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	summary, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check body composition
	if summary.BodyComposition == nil {
		t.Fatal("expected non-nil body_composition")
	}
	if w, ok := summary.BodyComposition["weight"]; !ok || w.Value != 185.4 {
		t.Errorf("expected weight 185.4, got %v", summary.BodyComposition["weight"])
	}
	if bf, ok := summary.BodyComposition["body_fat_pct"]; !ok || bf.Value != 22.3 {
		t.Errorf("expected body_fat_pct 22.3, got %v", summary.BodyComposition["body_fat_pct"])
	}

	// Check labs
	if summary.Labs == nil {
		t.Fatal("expected non-nil labs")
	}
	if len(summary.Labs.Markers) != 1 {
		t.Errorf("expected 1 lab marker, got %d", len(summary.Labs.Markers))
	}
	if summary.Labs.Markers[0].MeasurementType != "ldl_cholesterol" {
		t.Errorf("expected ldl_cholesterol marker, got %q", summary.Labs.Markers[0].MeasurementType)
	}

	// Check wearable
	if summary.Wearable == nil {
		t.Fatal("expected non-nil wearable")
	}
	if summary.Wearable.Source != "whoop" {
		t.Errorf("expected source 'whoop', got %q", summary.Wearable.Source)
	}
	if summary.Wearable.AvgHRV7d == nil || *summary.Wearable.AvgHRV7d != 45.2 {
		t.Errorf("expected avg_hrv_7d 45.2")
	}

	// Check nutrition
	if summary.Nutrition == nil {
		t.Fatal("expected non-nil nutrition")
	}
	if summary.Nutrition.AvgCalories7d == nil || *summary.Nutrition.AvgCalories7d != 2150 {
		t.Errorf("expected avg_calories_7d 2150")
	}
	if summary.Nutrition.AvgProtein7d == nil || *summary.Nutrition.AvgProtein7d != 165 {
		t.Errorf("expected avg_protein_7d 165")
	}
}

func TestGetClientProfileSummary_Unauthorized(t *testing.T) {
	ccRepo := &mockCoachClientRepo{
		findByCoachAndClient: func(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
			return nil, nil
		},
	}
	uc := newProfileSummaryUseCase(&mockMeasurementRepo{}, &mockWearableSummaryRepo{}, ccRepo, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err == nil {
		t.Fatal("expected authorization error")
	}
	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T", err)
	}
}

func TestGetClientProfileSummary_EmptyCoachID(t *testing.T) {
	uc := newProfileSummaryUseCase(&mockMeasurementRepo{}, &mockWearableSummaryRepo{}, &mockCoachClientRepo{}, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "", "client-1")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetClientProfileSummary_EmptyClientID(t *testing.T) {
	uc := newProfileSummaryUseCase(&mockMeasurementRepo{}, &mockWearableSummaryRepo{}, &mockCoachClientRepo{}, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "coach-1", "")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetClientProfileSummary_NoData(t *testing.T) {
	mRepo := &mockMeasurementRepo{
		findLatestByClientIDFunc: func(ctx context.Context, clientID string) ([]*entities.LatestMeasurement, error) {
			return []*entities.LatestMeasurement{}, nil
		},
	}
	wRepo := &mockWearableSummaryRepo{
		findAveragesFunc: func(ctx context.Context, clientID string, days int) (*entities.WearableAverages, error) {
			return nil, nil
		},
	}
	uc := newProfileSummaryUseCase(mRepo, wRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	summary, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.BodyComposition == nil {
		t.Fatal("expected non-nil body_composition map")
	}
	if len(summary.BodyComposition) != 0 {
		t.Errorf("expected empty body_composition, got %d entries", len(summary.BodyComposition))
	}
	if summary.Labs == nil {
		t.Fatal("expected non-nil labs")
	}
	if len(summary.Labs.Markers) != 0 {
		t.Errorf("expected 0 lab markers, got %d", len(summary.Labs.Markers))
	}
	if summary.Wearable != nil {
		t.Error("expected nil wearable when no data")
	}
	if summary.Nutrition != nil {
		t.Error("expected nil nutrition when no data")
	}
}

func TestGetClientProfileSummary_MeasurementRepoError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	mRepo := &mockMeasurementRepo{
		findLatestByClientIDFunc: func(ctx context.Context, clientID string) ([]*entities.LatestMeasurement, error) {
			return nil, repoErr
		},
	}
	uc := newProfileSummaryUseCase(mRepo, &mockWearableSummaryRepo{}, authorizedCoachClientRepo(), &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err == nil {
		t.Fatal("expected error from repository")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repository error to propagate, got %v", err)
	}
}

func TestGetClientProfileSummary_WearableRepoError(t *testing.T) {
	repoErr := errors.New("wearable db error")
	mRepo := &mockMeasurementRepo{
		findLatestByClientIDFunc: func(ctx context.Context, clientID string) ([]*entities.LatestMeasurement, error) {
			return []*entities.LatestMeasurement{}, nil
		},
	}
	wRepo := &mockWearableSummaryRepo{
		findAveragesFunc: func(ctx context.Context, clientID string, days int) (*entities.WearableAverages, error) {
			return nil, repoErr
		},
	}
	uc := newProfileSummaryUseCase(mRepo, wRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err == nil {
		t.Fatal("expected error from repository")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repository error to propagate, got %v", err)
	}
}

func TestGetClientProfileSummary_AuditEventLogged(t *testing.T) {
	var auditLogged bool
	auditRepo := &mockAuditRepo{
		logEventFunc: func(ctx context.Context, event *entities.AuditEvent) error {
			auditLogged = true
			if event.Action != "phi.access" {
				t.Errorf("expected action 'phi.access', got %q", event.Action)
			}
			if event.EntityType != "profile_summary" {
				t.Errorf("expected entity_type 'profile_summary', got %q", event.EntityType)
			}
			return nil
		},
	}
	mRepo := &mockMeasurementRepo{
		findLatestByClientIDFunc: func(ctx context.Context, clientID string) ([]*entities.LatestMeasurement, error) {
			return []*entities.LatestMeasurement{}, nil
		},
	}
	wRepo := &mockWearableSummaryRepo{
		findAveragesFunc: func(ctx context.Context, clientID string, days int) (*entities.WearableAverages, error) {
			return nil, nil
		},
	}
	uc := newProfileSummaryUseCase(mRepo, wRepo, authorizedCoachClientRepo(), auditRepo)

	_, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !auditLogged {
		t.Error("expected audit event to be logged")
	}
}
