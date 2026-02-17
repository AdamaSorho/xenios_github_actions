package handler

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

func TestNewLabExtractJobHandler_ValidPayload_Succeeds(t *testing.T) {
	artifactID := "art-test"
	csvData := []byte(`Test Name,Result,Units,Reference Range
Glucose,95,mg/dL,70-100
`)

	artifactRepo := repository.NewInMemoryArtifactRepository()
	artifact := &entities.Artifact{
		ID:         artifactID,
		ClientID:   "client-1",
		CoachID:    "coach-1",
		FileName:   "bloodwork.csv",
		Status:     entities.ArtifactStatusUploaded,
		StorageKey: "client-1/document/art-test.csv",
	}
	artifactRepo.Create(context.Background(), artifact)

	measurementRepo := repository.NewInMemoryMeasurementRepository()
	fileReader := repository.NewInMemoryFileContentReader()
	fileReader.PutObject("client-1/document/art-test.csv", csvData)

	auditRepo := repository.NewInMemoryAuditRepository()

	uc := usecase.NewExtractLabResultsUseCase(artifactRepo, measurementRepo, fileReader, auditRepo)
	handler := NewLabExtractJobHandler(uc)

	payload, _ := json.Marshal(LabExtractJobPayload{
		ArtifactID: artifactID,
		CoachID:    "coach-1",
	})

	job := &entities.Job{
		ID:      "job-1",
		Type:    entities.JobTypeExtractLabResults,
		Payload: payload,
	}

	err := handler(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify measurement was stored
	measurements, _ := measurementRepo.FindByArtifactID(context.Background(), artifactID)
	if len(measurements) != 1 {
		t.Errorf("expected 1 measurement stored, got %d", len(measurements))
	}
}

func TestNewLabExtractJobHandler_InvalidPayload_ReturnsError(t *testing.T) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	fileReader := repository.NewInMemoryFileContentReader()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := usecase.NewExtractLabResultsUseCase(artifactRepo, measurementRepo, fileReader, auditRepo)
	handler := NewLabExtractJobHandler(uc)

	job := &entities.Job{
		ID:      "job-2",
		Type:    entities.JobTypeExtractLabResults,
		Payload: []byte(`{invalid json`),
	}

	err := handler(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for invalid JSON payload")
	}
}

func TestNewLabExtractJobHandler_MissingArtifactID_ReturnsError(t *testing.T) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	fileReader := repository.NewInMemoryFileContentReader()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := usecase.NewExtractLabResultsUseCase(artifactRepo, measurementRepo, fileReader, auditRepo)
	handler := NewLabExtractJobHandler(uc)

	payload, _ := json.Marshal(LabExtractJobPayload{
		ArtifactID: "",
		CoachID:    "coach-1",
	})

	job := &entities.Job{
		ID:      "job-3",
		Type:    entities.JobTypeExtractLabResults,
		Payload: payload,
	}

	err := handler(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for missing artifact_id")
	}
}
