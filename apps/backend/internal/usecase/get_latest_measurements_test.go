package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newGetLatestMeasurementsUseCase() (*GetLatestMeasurementsUseCase, *repository.InMemoryMeasurementRepository, *repository.InMemoryCoachClientRepository, *repository.InMemoryAuditRepository) {
	mRepo := repository.NewInMemoryMeasurementRepository()
	ccRepo := repository.NewInMemoryCoachClientRepository()
	aRepo := repository.NewInMemoryAuditRepository()
	uc := NewGetLatestMeasurementsUseCase(mRepo, ccRepo, aRepo)
	return uc, mRepo, ccRepo, aRepo
}

func TestGetLatestMeasurements_ValidInput_ReturnsLatest(t *testing.T) {
	uc, mRepo, ccRepo, _ := newGetLatestMeasurementsUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	older := time.Date(2026, 1, 10, 10, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 1, 20, 10, 0, 0, 0, time.UTC)

	mRepo.Seed(&entities.Measurement{
		ID: "m-1", ClientID: "client-1", MeasurementType: "weight",
		Value: 185.4, Unit: "lbs", MeasuredAt: older,
	})
	mRepo.Seed(&entities.Measurement{
		ID: "m-2", ClientID: "client-1", MeasurementType: "weight",
		Value: 183.0, Unit: "lbs", MeasuredAt: newer,
	})
	mRepo.Seed(&entities.Measurement{
		ID: "m-3", ClientID: "client-1", MeasurementType: "body_fat_pct",
		Value: 22.3, Unit: "%", MeasuredAt: newer,
	})

	results, err := uc.Execute(context.Background(), GetLatestMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 latest measurements, got %d", len(results))
	}

	for _, r := range results {
		if r.MeasurementType == "weight" && r.Value != 183.0 {
			t.Errorf("expected latest weight 183.0, got %f", r.Value)
		}
	}
}

func TestGetLatestMeasurements_Unauthorized_ReturnsError(t *testing.T) {
	uc, _, _, _ := newGetLatestMeasurementsUseCase()

	_, err := uc.Execute(context.Background(), GetLatestMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})
	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T", err)
	}
}

func TestGetLatestMeasurements_EmptyData_ReturnsEmptySlice(t *testing.T) {
	uc, _, ccRepo, _ := newGetLatestMeasurementsUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	results, err := uc.Execute(context.Background(), GetLatestMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results == nil {
		t.Error("expected non-nil results")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestGetLatestMeasurements_MissingCoachID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newGetLatestMeasurementsUseCase()

	_, err := uc.Execute(context.Background(), GetLatestMeasurementsInput{
		CoachID:  "",
		ClientID: "client-1",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetLatestMeasurements_MissingClientID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newGetLatestMeasurementsUseCase()

	_, err := uc.Execute(context.Background(), GetLatestMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetLatestMeasurements_AuditEventLogged(t *testing.T) {
	uc, _, ccRepo, aRepo := newGetLatestMeasurementsUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	_, err := uc.Execute(context.Background(), GetLatestMeasurementsInput{
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
