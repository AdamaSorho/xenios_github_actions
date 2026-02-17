package handler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

func newTestInsightHandler() (*InsightGenerationHandler, *repository.InMemoryInsightCardRepository, *repository.InMemoryMeasurementRepository) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := usecase.NewGenerateInsightsUseCase(insightRepo, measurementRepo, auditRepo)
	handler := NewInsightGenerationHandler(uc)
	return handler, insightRepo, measurementRepo
}

func TestInsightGenerationHandler_Handle_ValidPayload_Succeeds(t *testing.T) {
	h, insightRepo, measurementRepo := newTestInsightHandler()

	refMax := 100.0
	measurementRepo.Add(&entities.Measurement{
		ID:           "m-1",
		ClientID:     "client-1",
		CoachID:      "coach-1",
		ArtifactID:   "artifact-1",
		Type:         entities.MeasurementTypeLab,
		MarkerName:   "LDL Cholesterol",
		Value:        142,
		Unit:         "mg/dL",
		ReferenceMax: &refMax,
		Flag:         entities.MeasurementFlagHigh,
		RecordedAt:   time.Now(),
		CreatedAt:    time.Now(),
	})

	payload, _ := json.Marshal(insightJobPayload{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
	})

	job := &entities.Job{
		ID:      "job-1",
		Type:    entities.JobTypeInsightGeneration,
		Payload: payload,
	}

	err := h.Handle(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if insightRepo.CardCount() != 1 {
		t.Errorf("expected 1 insight card, got %d", insightRepo.CardCount())
	}
}

func TestInsightGenerationHandler_Handle_InvalidJSON_ReturnsError(t *testing.T) {
	h, _, _ := newTestInsightHandler()

	job := &entities.Job{
		ID:      "job-2",
		Type:    entities.JobTypeInsightGeneration,
		Payload: []byte(`{invalid`),
	}

	err := h.Handle(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for invalid JSON payload")
	}
}

func TestInsightGenerationHandler_Handle_MissingFields_ReturnsError(t *testing.T) {
	h, _, _ := newTestInsightHandler()

	payload, _ := json.Marshal(insightJobPayload{
		ClientID: "client-1",
		// Missing CoachID and ArtifactID
	})

	job := &entities.Job{
		ID:      "job-3",
		Type:    entities.JobTypeInsightGeneration,
		Payload: payload,
	}

	err := h.Handle(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for missing required fields")
	}
}

func TestInsightGenerationHandler_Handle_NoMeasurements_Succeeds(t *testing.T) {
	h, insightRepo, _ := newTestInsightHandler()

	payload, _ := json.Marshal(insightJobPayload{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
	})

	job := &entities.Job{
		ID:      "job-4",
		Type:    entities.JobTypeInsightGeneration,
		Payload: payload,
	}

	err := h.Handle(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if insightRepo.CardCount() != 0 {
		t.Errorf("expected 0 insight cards, got %d", insightRepo.CardCount())
	}
}
