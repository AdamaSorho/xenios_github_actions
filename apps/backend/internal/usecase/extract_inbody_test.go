package usecase

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
	domainrepo "github.com/xenios/backend/internal/domain/repository"
	"github.com/xenios/backend/internal/infrastructure/pdfextract"
)

func newExtractInBodyUseCase() (*ExtractInBodyUseCase, *repository.InMemoryArtifactRepository, *repository.InMemoryFileStorage, *repository.InMemoryMeasurementRepository, *repository.InMemoryAuditRepository) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	auditRepo := repository.NewInMemoryAuditRepository()
	extractor := pdfextract.NewInBodyExtractor()

	uc := NewExtractInBodyUseCase(artifactRepo, fileStorage, measurementRepo, auditRepo, extractor)
	return uc, artifactRepo, fileStorage, measurementRepo, auditRepo
}

func createUploadedInBodyArtifact(t *testing.T, repo *repository.InMemoryArtifactRepository, clientID, coachID string) *entities.Artifact {
	t.Helper()
	art := &entities.Artifact{
		ClientID:    clientID,
		CoachID:     coachID,
		FileName:    "inbody_scan.pdf",
		FileType:    "application/pdf",
		FileSize:    2048,
		StorageKey:  clientID + "/document/inbody-scan.pdf",
		Type:        entities.ArtifactTypeDocument,
		Status:      entities.ArtifactStatusUploaded,
		ContentType: "application/pdf",
	}
	created, err := repo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("failed to create test artifact: %v", err)
	}
	return created
}

func sampleInBodyPDFContent() []byte {
	return []byte(`
InBody 570 Body Composition Analysis

Weight: 85.4 kg
Skeletal Muscle Mass: 38.2 kg
Body Fat Percentage: 22.3%
Total Body Water: 42.1 L
Lean Body Mass: 66.5 kg
Basal Metabolic Rate: 1847 kcal
`)
}

func samplePartialPDFContent() []byte {
	return []byte(`
InBody Report
Weight: 85.4 kg
Body Fat Percentage: 22.3%
`)
}

func TestExtractInBody_ValidInput_ExtractsAndStoresMeasurements(t *testing.T) {
	uc, artifactRepo, fileStorage, measurementRepo, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo, "client-1", "coach-1")

	fileStorage.PutObjectWithContent(artifact.StorageKey, sampleInBodyPDFContent())

	input := ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	}

	out, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.ArtifactID != artifact.ID {
		t.Errorf("expected artifact_id %q, got %q", artifact.ID, out.ArtifactID)
	}
	if out.Partial {
		t.Errorf("expected full extraction, got partial; errors: %v", out.Errors)
	}
	if len(out.Measurements) != 6 {
		t.Fatalf("expected 6 measurements, got %d", len(out.Measurements))
	}

	// Verify measurements are stored in repository
	stored, err := measurementRepo.FindByArtifactID(context.Background(), artifact.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stored) != 6 {
		t.Errorf("expected 6 stored measurements, got %d", len(stored))
	}

	// Verify each measurement has correct metadata
	for _, m := range stored {
		if m.ClientID != "client-1" {
			t.Errorf("expected client_id 'client-1', got %q", m.ClientID)
		}
		if m.RecordedBy != "coach-1" {
			t.Errorf("expected recorded_by 'coach-1', got %q", m.RecordedBy)
		}
		if m.ArtifactID != artifact.ID {
			t.Errorf("expected artifact_id %q, got %q", artifact.ID, m.ArtifactID)
		}
	}
}

func TestExtractInBody_EmptyArtifactID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newExtractInBodyUseCase()

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: "",
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractInBody_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newExtractInBodyUseCase()

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: "art-1",
		CoachID:    "",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractInBody_ArtifactNotFound_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newExtractInBodyUseCase()

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: "nonexistent",
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractInBody_WrongCoach_ReturnsAuthError(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo, "client-1", "coach-1")
	fileStorage.PutObjectWithContent(artifact.StorageKey, sampleInBodyPDFContent())

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "different-coach",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func TestExtractInBody_ArtifactNotUploaded_ReturnsValidationError(t *testing.T) {
	uc, artifactRepo, _, _, _ := newExtractInBodyUseCase()

	// Create a pending artifact (not yet uploaded)
	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "inbody_scan.pdf",
		FileType:    "application/pdf",
		FileSize:    2048,
		StorageKey:  "client-1/document/pending.pdf",
		Type:        entities.ArtifactTypeDocument,
		Status:      entities.ArtifactStatusPending,
		ContentType: "application/pdf",
	}
	created, _ := artifactRepo.Create(context.Background(), art)

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: created.ID,
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractInBody_CorruptPDF_ReturnsError(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo, "client-1", "coach-1")

	fileStorage.PutObjectWithContent(artifact.StorageKey, []byte("not an inbody pdf content at all"))

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error for corrupt PDF")
	}
}

func TestExtractInBody_PartialExtraction_ReturnsMeasurementsAndFlags(t *testing.T) {
	uc, artifactRepo, fileStorage, measurementRepo, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo, "client-1", "coach-1")

	fileStorage.PutObjectWithContent(artifact.StorageKey, samplePartialPDFContent())

	out, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !out.Partial {
		t.Error("expected partial extraction flag")
	}

	if len(out.Measurements) != 2 {
		t.Errorf("expected 2 measurements, got %d", len(out.Measurements))
	}

	if len(out.Errors) == 0 {
		t.Error("expected error details for missing fields")
	}

	// Verify partial measurements are still stored
	stored, err := measurementRepo.FindByArtifactID(context.Background(), artifact.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stored) != 2 {
		t.Errorf("expected 2 stored measurements, got %d", len(stored))
	}
}

func TestExtractInBody_Success_AuditEventLogged(t *testing.T) {
	uc, artifactRepo, fileStorage, _, auditRepo := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo, "client-1", "coach-1")
	fileStorage.PutObjectWithContent(artifact.StorageKey, sampleInBodyPDFContent())

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	if len(events) == 0 {
		t.Fatal("expected audit event to be logged")
	}

	found := false
	for _, e := range events {
		if e.Action == "artifact.extraction_success" {
			found = true
			if e.EntityID != artifact.ID {
				t.Errorf("expected entity_id %q, got %q", artifact.ID, e.EntityID)
			}
			break
		}
	}
	if !found {
		t.Error("expected 'artifact.extraction_success' audit event")
	}
}

func TestExtractInBody_PartialExtraction_AuditEventLogged(t *testing.T) {
	uc, artifactRepo, fileStorage, _, auditRepo := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo, "client-1", "coach-1")
	fileStorage.PutObjectWithContent(artifact.StorageKey, samplePartialPDFContent())

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	found := false
	for _, e := range events {
		if e.Action == "artifact.extraction_partial" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'artifact.extraction_partial' audit event")
	}
}

func TestExtractInBody_FailedExtraction_AuditEventLogged(t *testing.T) {
	uc, artifactRepo, fileStorage, _, auditRepo := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo, "client-1", "coach-1")
	fileStorage.PutObjectWithContent(artifact.StorageKey, []byte("garbage data no metrics"))

	_, _ = uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})

	events := auditRepo.GetEvents()
	found := false
	for _, e := range events {
		if e.Action == "artifact.extraction_failed" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'artifact.extraction_failed' audit event")
	}
}

func TestExtractInBody_MeasuredAt_SetFromInput(t *testing.T) {
	uc, artifactRepo, fileStorage, measurementRepo, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo, "client-1", "coach-1")
	fileStorage.PutObjectWithContent(artifact.StorageKey, sampleInBodyPDFContent())

	measuredAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	out, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
		MeasuredAt: &measuredAt,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, m := range out.Measurements {
		if !m.MeasuredAt.Equal(measuredAt) {
			t.Errorf("expected measured_at %v, got %v", measuredAt, m.MeasuredAt)
		}
	}

	stored, _ := measurementRepo.FindByArtifactID(context.Background(), artifact.ID)
	for _, m := range stored {
		if !m.MeasuredAt.Equal(measuredAt) {
			t.Errorf("stored measured_at: expected %v, got %v", measuredAt, m.MeasuredAt)
		}
	}
}

func TestExtractInBody_FromJob_ParsesPayload(t *testing.T) {
	payload := ExtractInBodyJobPayload{
		ArtifactID: "art-123",
		CoachID:    "coach-1",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed ExtractInBodyJobPayload
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.ArtifactID != "art-123" {
		t.Errorf("expected artifact_id 'art-123', got %q", parsed.ArtifactID)
	}
	if parsed.CoachID != "coach-1" {
		t.Errorf("expected coach_id 'coach-1', got %q", parsed.CoachID)
	}
}

func TestExtractInBody_FileNotInStorage_ReturnsError(t *testing.T) {
	uc, artifactRepo, _, _, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo, "client-1", "coach-1")

	// Don't put the file in storage
	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error when file not in storage")
	}
}

func TestNewExtractInBodyJobHandler_ProcessesJob(t *testing.T) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	auditRepo := repository.NewInMemoryAuditRepository()
	extractor := pdfextract.NewInBodyExtractor()

	uc := NewExtractInBodyUseCase(artifactRepo, fileStorage, measurementRepo, auditRepo, extractor)
	handler := NewExtractInBodyJobHandler(uc)

	artifact := createUploadedInBodyArtifact(t, artifactRepo, "client-1", "coach-1")
	fileStorage.PutObjectWithContent(artifact.StorageKey, sampleInBodyPDFContent())

	payload, _ := json.Marshal(ExtractInBodyJobPayload{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})

	job := &entities.Job{
		ID:      "job-1",
		Type:    entities.JobTypeExtractInBody,
		Payload: payload,
		Status:  entities.JobStatusActive,
		Attempt: 1,
	}

	err := handler(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, _ := measurementRepo.FindByArtifactID(context.Background(), artifact.ID)
	if len(stored) != 6 {
		t.Errorf("expected 6 measurements, got %d", len(stored))
	}
}

func TestNewExtractInBodyJobHandler_InvalidPayload_ReturnsError(t *testing.T) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	auditRepo := repository.NewInMemoryAuditRepository()
	extractor := pdfextract.NewInBodyExtractor()

	uc := NewExtractInBodyUseCase(artifactRepo, fileStorage, measurementRepo, auditRepo, extractor)
	handler := NewExtractInBodyJobHandler(uc)

	job := &entities.Job{
		ID:      "job-2",
		Type:    entities.JobTypeExtractInBody,
		Payload: []byte(`{invalid json`),
		Status:  entities.JobStatusActive,
		Attempt: 1,
	}

	err := handler(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for invalid payload")
	}
}

// Verify the PDFExtractor interface conformance
var _ domainrepo.PDFExtractor = (*pdfextract.InBodyExtractor)(nil)
