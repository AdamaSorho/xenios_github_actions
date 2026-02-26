package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newGetClientMeasurementsUseCase() (*GetClientMeasurementsUseCase, *repository.InMemoryMeasurementRepository, *repository.InMemoryCoachClientRepository, *repository.InMemoryAuditRepository) {
	mRepo := repository.NewInMemoryMeasurementRepository()
	ccRepo := repository.NewInMemoryCoachClientRepository()
	aRepo := repository.NewInMemoryAuditRepository()
	uc := NewGetClientMeasurementsUseCase(mRepo, ccRepo, aRepo)
	return uc, mRepo, ccRepo, aRepo
}

func seedCoachClient(t *testing.T, ccRepo *repository.InMemoryCoachClientRepository, coachID, clientID string) {
	t.Helper()
	_, err := ccRepo.Create(context.Background(), coachID, clientID)
	if err != nil {
		t.Fatalf("failed to seed coach-client: %v", err)
	}
}

func TestGetClientMeasurements_ValidInput_ReturnsMeasurements(t *testing.T) {
	uc, mRepo, ccRepo, _ := newGetClientMeasurementsUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	mRepo.Seed(&entities.Measurement{
		ID: "m-1", ClientID: "client-1", MeasurementType: "weight",
		Value: 185.4, Unit: "lbs", MeasuredAt: time.Now(),
	})
	mRepo.Seed(&entities.Measurement{
		ID: "m-2", ClientID: "client-1", MeasurementType: "body_fat_pct",
		Value: 22.3, Unit: "%", MeasuredAt: time.Now(),
	})

	result, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Filter:   entities.MeasurementFilter{Page: 1, Limit: 20},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Measurements) != 2 {
		t.Errorf("expected 2 measurements, got %d", len(result.Measurements))
	}
	if result.Pagination.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Pagination.Total)
	}
}

func TestGetClientMeasurements_FilterByType_ReturnsFiltered(t *testing.T) {
	uc, mRepo, ccRepo, _ := newGetClientMeasurementsUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	mRepo.Seed(&entities.Measurement{
		ID: "m-1", ClientID: "client-1", MeasurementType: "weight",
		Value: 185.4, Unit: "lbs", MeasuredAt: time.Now(),
	})
	mRepo.Seed(&entities.Measurement{
		ID: "m-2", ClientID: "client-1", MeasurementType: "body_fat_pct",
		Value: 22.3, Unit: "%", MeasuredAt: time.Now(),
	})

	result, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Filter:   entities.MeasurementFilter{MeasurementType: "weight", Page: 1, Limit: 20},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Measurements) != 1 {
		t.Errorf("expected 1 measurement, got %d", len(result.Measurements))
	}
	if result.Measurements[0].MeasurementType != "weight" {
		t.Errorf("expected type weight, got %s", result.Measurements[0].MeasurementType)
	}
}

func TestGetClientMeasurements_FilterByDateRange_ReturnsInRange(t *testing.T) {
	uc, mRepo, ccRepo, _ := newGetClientMeasurementsUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	jan15 := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	feb15 := time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC)
	mRepo.Seed(&entities.Measurement{
		ID: "m-1", ClientID: "client-1", MeasurementType: "weight",
		Value: 185.4, Unit: "lbs", MeasuredAt: jan15,
	})
	mRepo.Seed(&entities.Measurement{
		ID: "m-2", ClientID: "client-1", MeasurementType: "weight",
		Value: 183.0, Unit: "lbs", MeasuredAt: feb15,
	})

	from := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 2, 28, 23, 59, 59, 0, time.UTC)
	result, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Filter:   entities.MeasurementFilter{From: &from, To: &to, Page: 1, Limit: 20},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Measurements) != 1 {
		t.Errorf("expected 1 measurement in range, got %d", len(result.Measurements))
	}
}

func TestGetClientMeasurements_Pagination_ReturnsCorrectPage(t *testing.T) {
	uc, mRepo, ccRepo, _ := newGetClientMeasurementsUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	for i := 0; i < 25; i++ {
		mRepo.Seed(&entities.Measurement{
			ID: "m-" + string(rune('a'+i)), ClientID: "client-1", MeasurementType: "weight",
			Value: float64(180 + i), Unit: "lbs", MeasuredAt: time.Now().Add(time.Duration(-i) * time.Hour),
		})
	}

	result, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Filter:   entities.MeasurementFilter{Page: 2, Limit: 10},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Measurements) != 10 {
		t.Errorf("expected 10 measurements on page 2, got %d", len(result.Measurements))
	}
	if result.Pagination.Page != 2 {
		t.Errorf("expected page 2, got %d", result.Pagination.Page)
	}
	if result.Pagination.Total != 25 {
		t.Errorf("expected total 25, got %d", result.Pagination.Total)
	}
}

func TestGetClientMeasurements_Unauthorized_ReturnsError(t *testing.T) {
	uc, _, _, _ := newGetClientMeasurementsUseCase()

	_, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Filter:   entities.MeasurementFilter{Page: 1, Limit: 20},
	})
	if err == nil {
		t.Fatal("expected authorization error")
	}
	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T: %v", err, err)
	}
}

func TestGetClientMeasurements_EmptyData_ReturnsEmptySlice(t *testing.T) {
	uc, _, ccRepo, _ := newGetClientMeasurementsUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	result, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Filter:   entities.MeasurementFilter{Page: 1, Limit: 20},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Measurements == nil {
		t.Error("expected non-nil measurements slice")
	}
	if len(result.Measurements) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(result.Measurements))
	}
}

func TestGetClientMeasurements_MissingCoachID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newGetClientMeasurementsUseCase()

	_, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "",
		ClientID: "client-1",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetClientMeasurements_MissingClientID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newGetClientMeasurementsUseCase()

	_, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetClientMeasurements_AuditEventLogged(t *testing.T) {
	uc, _, ccRepo, aRepo := newGetClientMeasurementsUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	_, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Filter:   entities.MeasurementFilter{Page: 1, Limit: 20},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if aRepo.EventCount() < 1 {
		t.Error("expected at least 1 audit event")
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

func TestGetClientMeasurements_DefaultPagination(t *testing.T) {
	uc, _, ccRepo, _ := newGetClientMeasurementsUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	result, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Filter:   entities.MeasurementFilter{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Pagination.Page != 1 {
		t.Errorf("expected default page 1, got %d", result.Pagination.Page)
	}
	if result.Pagination.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", result.Pagination.Limit)
	}
}
