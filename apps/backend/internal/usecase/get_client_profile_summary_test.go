package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newGetClientProfileSummaryUseCase() (*GetClientProfileSummaryUseCase, *repository.InMemoryMeasurementRepository, *repository.InMemoryWearableSummaryRepository, *repository.InMemoryCoachClientRepository, *repository.InMemoryAuditRepository) {
	mRepo := repository.NewInMemoryMeasurementRepository()
	wRepo := repository.NewInMemoryWearableSummaryRepository()
	ccRepo := repository.NewInMemoryCoachClientRepository()
	aRepo := repository.NewInMemoryAuditRepository()
	uc := NewGetClientProfileSummaryUseCase(mRepo, wRepo, ccRepo, aRepo)
	return uc, mRepo, wRepo, ccRepo, aRepo
}

func TestGetClientProfileSummary_ValidInput_ReturnsSummary(t *testing.T) {
	uc, mRepo, wRepo, ccRepo, _ := newGetClientProfileSummaryUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	now := time.Now()
	highFlag := "high"

	mRepo.Seed(&entities.Measurement{
		ID: "m-1", ClientID: "client-1", MeasurementType: "weight",
		Value: 185.4, Unit: "lbs", MeasuredAt: now,
	})
	mRepo.Seed(&entities.Measurement{
		ID: "m-2", ClientID: "client-1", MeasurementType: "body_fat_pct",
		Value: 22.3, Unit: "%", MeasuredAt: now,
	})
	mRepo.Seed(&entities.Measurement{
		ID: "m-3", ClientID: "client-1", MeasurementType: "ldl_cholesterol",
		Value: 142, Unit: "mg/dL", MeasuredAt: now, Flag: &highFlag,
	})
	mRepo.Seed(&entities.Measurement{
		ID: "m-4", ClientID: "client-1", MeasurementType: "calories",
		Value: 2150, Unit: "kcal", MeasuredAt: now,
	})
	mRepo.Seed(&entities.Measurement{
		ID: "m-5", ClientID: "client-1", MeasurementType: "protein",
		Value: 165, Unit: "g", MeasuredAt: now,
	})

	wRepo.Seed(&entities.WearableSummary{
		ID: "ws-1", ClientID: "client-1", Source: "whoop",
		SummaryDate: "2026-01-15",
		Metrics:     map[string]interface{}{"hrv": 45.2, "sleep_hours": 7.2, "recovery_score": 68.0},
	})

	result, err := uc.Execute(context.Background(), GetClientProfileSummaryInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.BodyComposition["weight"] == nil {
		t.Error("expected weight in body composition")
	}
	if result.BodyComposition["weight"].Value != 185.4 {
		t.Errorf("expected weight 185.4, got %f", result.BodyComposition["weight"].Value)
	}

	if result.Labs.FlaggedCount != 1 {
		t.Errorf("expected 1 flagged lab, got %d", result.Labs.FlaggedCount)
	}
	if len(result.Labs.Markers) != 1 {
		t.Errorf("expected 1 lab marker, got %d", len(result.Labs.Markers))
	}

	if result.Wearable.AvgHRV7d == nil || *result.Wearable.AvgHRV7d != 45.2 {
		t.Errorf("expected avg HRV 45.2, got %v", result.Wearable.AvgHRV7d)
	}
	if result.Wearable.Source == nil || *result.Wearable.Source != "whoop" {
		t.Error("expected wearable source whoop")
	}

	if result.Nutrition.AvgCalories7d == nil || *result.Nutrition.AvgCalories7d != 2150 {
		t.Errorf("expected avg calories 2150, got %v", result.Nutrition.AvgCalories7d)
	}
	if result.Nutrition.AvgProtein7d == nil || *result.Nutrition.AvgProtein7d != 165 {
		t.Errorf("expected avg protein 165, got %v", result.Nutrition.AvgProtein7d)
	}
}

func TestGetClientProfileSummary_Unauthorized_ReturnsError(t *testing.T) {
	uc, _, _, _, _ := newGetClientProfileSummaryUseCase()

	_, err := uc.Execute(context.Background(), GetClientProfileSummaryInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})
	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T", err)
	}
}

func TestGetClientProfileSummary_EmptyData_ReturnsEmptyProfile(t *testing.T) {
	uc, _, _, ccRepo, _ := newGetClientProfileSummaryUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	result, err := uc.Execute(context.Background(), GetClientProfileSummaryInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.BodyComposition) != 0 {
		t.Errorf("expected empty body composition, got %d entries", len(result.BodyComposition))
	}
	if result.Labs.FlaggedCount != 0 {
		t.Errorf("expected 0 flagged, got %d", result.Labs.FlaggedCount)
	}
	if len(result.Labs.Markers) != 0 {
		t.Errorf("expected 0 lab markers, got %d", len(result.Labs.Markers))
	}
}

func TestGetClientProfileSummary_MissingCoachID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newGetClientProfileSummaryUseCase()

	_, err := uc.Execute(context.Background(), GetClientProfileSummaryInput{
		CoachID:  "",
		ClientID: "client-1",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetClientProfileSummary_MissingClientID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newGetClientProfileSummaryUseCase()

	_, err := uc.Execute(context.Background(), GetClientProfileSummaryInput{
		CoachID:  "coach-1",
		ClientID: "",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetClientProfileSummary_AuditEventLogged(t *testing.T) {
	uc, _, _, ccRepo, aRepo := newGetClientProfileSummaryUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	_, err := uc.Execute(context.Background(), GetClientProfileSummaryInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	events := aRepo.GetEvents()
	found := false
	for _, e := range events {
		if e.Action == "phi.access" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected phi.access audit event")
	}
}

func TestGetClientProfileSummary_MultipleWearables_ComputesAverages(t *testing.T) {
	uc, _, wRepo, ccRepo, _ := newGetClientProfileSummaryUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	wRepo.Seed(&entities.WearableSummary{
		ID: "ws-1", ClientID: "client-1", Source: "whoop",
		SummaryDate: "2026-01-15",
		Metrics:     map[string]interface{}{"hrv": 40.0, "sleep_hours": 7.0, "recovery_score": 60.0},
	})
	wRepo.Seed(&entities.WearableSummary{
		ID: "ws-2", ClientID: "client-1", Source: "whoop",
		SummaryDate: "2026-01-14",
		Metrics:     map[string]interface{}{"hrv": 50.0, "sleep_hours": 8.0, "recovery_score": 80.0},
	})

	result, err := uc.Execute(context.Background(), GetClientProfileSummaryInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Wearable.AvgHRV7d == nil || *result.Wearable.AvgHRV7d != 45.0 {
		t.Errorf("expected avg HRV 45.0, got %v", result.Wearable.AvgHRV7d)
	}
	if result.Wearable.AvgSleep7d == nil || *result.Wearable.AvgSleep7d != 7.5 {
		t.Errorf("expected avg sleep 7.5, got %v", result.Wearable.AvgSleep7d)
	}
	if result.Wearable.AvgRecovery7d == nil || *result.Wearable.AvgRecovery7d != 70.0 {
		t.Errorf("expected avg recovery 70.0, got %v", result.Wearable.AvgRecovery7d)
	}
}

func TestExtractFloat_IntTypes(t *testing.T) {
	m := map[string]interface{}{
		"int_val":   42,
		"int64_val": int64(100),
		"float_val": 3.14,
		"str_val":   "not a number",
	}

	if v, ok := extractFloat(m, "int_val"); !ok || v != 42.0 {
		t.Errorf("expected 42.0, got %v (ok=%v)", v, ok)
	}
	if v, ok := extractFloat(m, "int64_val"); !ok || v != 100.0 {
		t.Errorf("expected 100.0, got %v (ok=%v)", v, ok)
	}
	if v, ok := extractFloat(m, "float_val"); !ok || v != 3.14 {
		t.Errorf("expected 3.14, got %v (ok=%v)", v, ok)
	}
	if _, ok := extractFloat(m, "str_val"); ok {
		t.Error("expected false for string value")
	}
	if _, ok := extractFloat(m, "missing"); ok {
		t.Error("expected false for missing key")
	}
}
