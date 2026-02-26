package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newGetWearableSummariesUseCase() (*GetWearableSummariesUseCase, *repository.InMemoryWearableSummaryRepository, *repository.InMemoryCoachClientRepository, *repository.InMemoryAuditRepository) {
	wRepo := repository.NewInMemoryWearableSummaryRepository()
	ccRepo := repository.NewInMemoryCoachClientRepository()
	aRepo := repository.NewInMemoryAuditRepository()
	uc := NewGetWearableSummariesUseCase(wRepo, ccRepo, aRepo)
	return uc, wRepo, ccRepo, aRepo
}

func TestGetWearableSummaries_ValidInput_ReturnsSummaries(t *testing.T) {
	uc, wRepo, ccRepo, _ := newGetWearableSummariesUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	wRepo.Seed(&entities.WearableSummary{
		ID: "ws-1", ClientID: "client-1", Source: "whoop",
		SummaryDate: "2026-01-15",
		Metrics:     map[string]interface{}{"hrv": 45.2, "sleep_hours": 7.5},
	})
	wRepo.Seed(&entities.WearableSummary{
		ID: "ws-2", ClientID: "client-1", Source: "whoop",
		SummaryDate: "2026-01-14",
		Metrics:     map[string]interface{}{"hrv": 42.0, "sleep_hours": 6.8},
	})

	results, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Limit:    30,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 summaries, got %d", len(results))
	}
}

func TestGetWearableSummaries_Unauthorized_ReturnsError(t *testing.T) {
	uc, _, _, _ := newGetWearableSummariesUseCase()

	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})
	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T", err)
	}
}

func TestGetWearableSummaries_EmptyData_ReturnsEmptySlice(t *testing.T) {
	uc, _, ccRepo, _ := newGetWearableSummariesUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	results, err := uc.Execute(context.Background(), GetWearableSummariesInput{
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

func TestGetWearableSummaries_DefaultLimit(t *testing.T) {
	uc, _, ccRepo, _ := newGetWearableSummariesUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Limit:    0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetWearableSummaries_MissingCoachID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newGetWearableSummariesUseCase()

	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "",
		ClientID: "client-1",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetWearableSummaries_MissingClientID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newGetWearableSummariesUseCase()

	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetWearableSummaries_AuditEventLogged(t *testing.T) {
	uc, _, ccRepo, aRepo := newGetWearableSummariesUseCase()
	seedCoachClient(t, ccRepo, "coach-1", "client-1")

	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
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
