package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

func newTestInsightHandler() (JobHandler, *repository.InMemoryInsightCardRepository, *repository.InMemoryMeasurementRepository) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	measureRepo := repository.NewInMemoryMeasurementRepository()
	wearableRepo := repository.NewInMemoryWearableSummaryRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := usecase.NewGenerateInsightsUseCase(insightRepo, measureRepo, wearableRepo, auditRepo)
	handler := NewInsightGenerationHandler(uc)
	return handler, insightRepo, measureRepo
}

func TestInsightGenerationHandler_ValidPayload_GeneratesInsights(t *testing.T) {
	handler, insightRepo, measureRepo := newTestInsightHandler()

	now := time.Now()
	measureRepo.Add(&entities.Measurement{
		ID:              "m-1",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "ldl_cholesterol",
		Value:           142,
		Unit:            "mg/dL",
		MeasuredAt:      now,
		CreatedAt:       now,
	})

	payload, _ := json.Marshal(InsightJobPayload{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "art-1",
	})

	job := &entities.Job{
		ID:      "job-1",
		Type:    entities.JobTypeInsightGeneration,
		Payload: payload,
	}

	err := handler(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cards := insightRepo.GetAll()
	if len(cards) != 1 {
		t.Errorf("expected 1 insight card, got %d", len(cards))
	}
}

func TestInsightGenerationHandler_InvalidJSON_ReturnsError(t *testing.T) {
	handler, _, _ := newTestInsightHandler()

	job := &entities.Job{
		ID:      "job-2",
		Type:    entities.JobTypeInsightGeneration,
		Payload: []byte(`{invalid}`),
	}

	err := handler(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestInsightGenerationHandler_MissingClientID_ReturnsError(t *testing.T) {
	handler, _, _ := newTestInsightHandler()

	payload, _ := json.Marshal(InsightJobPayload{
		CoachID: "coach-1",
	})

	job := &entities.Job{
		ID:      "job-3",
		Type:    entities.JobTypeInsightGeneration,
		Payload: payload,
	}

	err := handler(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for missing client_id")
	}
}

func TestInsightGenerationHandler_MissingCoachID_ReturnsError(t *testing.T) {
	handler, _, _ := newTestInsightHandler()

	payload, _ := json.Marshal(InsightJobPayload{
		ClientID: "client-1",
	})

	job := &entities.Job{
		ID:      "job-4",
		Type:    entities.JobTypeInsightGeneration,
		Payload: payload,
	}

	err := handler(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for missing coach_id")
	}
}

func TestInsightGenerationHandler_NoData_Succeeds(t *testing.T) {
	handler, _, _ := newTestInsightHandler()

	payload, _ := json.Marshal(InsightJobPayload{
		ClientID: "client-1",
		CoachID:  "coach-1",
	})

	job := &entities.Job{
		ID:      "job-5",
		Type:    entities.JobTypeInsightGeneration,
		Payload: payload,
	}

	err := handler(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
