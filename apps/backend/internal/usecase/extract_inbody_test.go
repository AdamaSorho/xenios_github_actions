package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

const sampleInBodyText = `InBody 570 Body Composition Analysis

Weight: 185.4 lbs
Skeletal Muscle Mass: 78.2 lbs
Body Fat Percentage: 22.3 %
Basal Metabolic Rate: 1847 kcal
Total Body Water: 42.1 L
Lean Body Mass: 144.1 lbs`

const partialInBodyText = `Weight: 185.4 lbs
Body Fat: 22.3 %`

func newExtractInBodyTestDeps() (
	*ExtractInBodyUseCase,
	*repository.InMemoryArtifactRepository,
	*repository.InMemoryFileDownloader,
	*repository.InMemoryPDFTextExtractor,
	*repository.InMemoryMeasurementRepository,
	*repository.InMemoryAuditRepository,
) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileDownloader := repository.NewInMemoryFileDownloader()
	pdfExtractor := repository.NewInMemoryPDFTextExtractor()
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := NewExtractInBodyUseCase(artifactRepo, fileDownloader, pdfExtractor, measurementRepo, auditRepo)
	return uc, artifactRepo, fileDownloader, pdfExtractor, measurementRepo, auditRepo
}

func createTestArtifact(t *testing.T, artifactRepo *repository.InMemoryArtifactRepository) *entities.Artifact {
	t.Helper()
	artifact, err := artifactRepo.Create(context.Background(), &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "inbody_scan.pdf",
		FileType:    "application/pdf",
		FileSize:    5000,
		StorageKey:  "client-1/document/test-artifact.pdf",
		Type:        entities.ArtifactTypeDocument,
		Status:      entities.ArtifactStatusUploaded,
		ContentType: "application/pdf",
	})
	if err != nil {
		t.Fatalf("failed to create test artifact: %v", err)
	}
	return artifact
}

func TestExtractInBody_HappyPath_ExtractsAllMetrics(t *testing.T) {
	uc, artifactRepo, fileDownloader, pdfExtractor, measurementRepo, auditRepo := newExtractInBodyTestDeps()
	artifact := createTestArtifact(t, artifactRepo)

	pdfContent := "fake-pdf-content"
	fileDownloader.PutFile(artifact.StorageKey, []byte(pdfContent))
	pdfExtractor.SetTextForContent(pdfContent, sampleInBodyText)

	input := ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.MeasurementsCreated != 6 {
		t.Errorf("expected 6 measurements, got %d", output.MeasurementsCreated)
	}
	if output.Partial {
		t.Error("expected full extraction, got partial")
	}

	// Verify measurements stored
	measurements := measurementRepo.GetAll()
	if len(measurements) != 6 {
		t.Fatalf("expected 6 stored measurements, got %d", len(measurements))
	}

	// Verify all measurements reference the artifact
	for _, m := range measurements {
		if m.ArtifactID == nil || *m.ArtifactID != artifact.ID {
			t.Errorf("expected artifact_id %q, got %v", artifact.ID, m.ArtifactID)
		}
		if m.ClientID != "client-1" {
			t.Errorf("expected client_id 'client-1', got %q", m.ClientID)
		}
	}

	// Verify artifact status updated to processed
	updated, _ := artifactRepo.FindByID(context.Background(), artifact.ID)
	if updated.Status != entities.ArtifactStatusProcessed {
		t.Errorf("expected artifact status 'processed', got %q", updated.Status)
	}

	// Verify audit event logged
	events := auditRepo.GetEvents()
	foundSuccess := false
	for _, e := range events {
		if e.Action == "inbody.extraction_success" {
			foundSuccess = true
		}
	}
	if !foundSuccess {
		t.Error("expected audit event 'inbody.extraction_success'")
	}
}

func TestExtractInBody_PartialExtraction_FlagsPartialAndStoresExtracted(t *testing.T) {
	uc, artifactRepo, fileDownloader, pdfExtractor, measurementRepo, _ := newExtractInBodyTestDeps()
	artifact := createTestArtifact(t, artifactRepo)

	pdfContent := "partial-pdf"
	fileDownloader.PutFile(artifact.StorageKey, []byte(pdfContent))
	pdfExtractor.SetTextForContent(pdfContent, partialInBodyText)

	output, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Partial {
		t.Error("expected partial flag to be set")
	}
	if output.MeasurementsCreated != 2 {
		t.Errorf("expected 2 measurements, got %d", output.MeasurementsCreated)
	}

	measurements := measurementRepo.GetAll()
	if len(measurements) != 2 {
		t.Errorf("expected 2 stored measurements, got %d", len(measurements))
	}

	// Partial results should still update artifact to processed
	updated, _ := artifactRepo.FindByID(context.Background(), artifact.ID)
	if updated.Status != entities.ArtifactStatusProcessed {
		t.Errorf("expected artifact status 'processed', got %q", updated.Status)
	}
}

func TestExtractInBody_EmptyArtifactID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _, _ := newExtractInBodyTestDeps()

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
	uc, _, _, _, _, _ := newExtractInBodyTestDeps()

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: "some-id",
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
	uc, _, _, _, _, _ := newExtractInBodyTestDeps()

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
	uc, artifactRepo, _, _, _, _ := newExtractInBodyTestDeps()
	artifact := createTestArtifact(t, artifactRepo)

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "wrong-coach",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T: %v", err, err)
	}
}

func TestExtractInBody_ArtifactNotUploaded_ReturnsValidationError(t *testing.T) {
	uc, artifactRepo, _, _, _, _ := newExtractInBodyTestDeps()

	artifact, _ := artifactRepo.Create(context.Background(), &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "inbody.pdf",
		FileType:    "application/pdf",
		FileSize:    5000,
		StorageKey:  "client-1/document/pending.pdf",
		Type:        entities.ArtifactTypeDocument,
		Status:      entities.ArtifactStatusPending, // not uploaded
		ContentType: "application/pdf",
	})

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractInBody_DownloadFails_ReturnsError(t *testing.T) {
	uc, artifactRepo, _, _, _, auditRepo := newExtractInBodyTestDeps()
	artifact := createTestArtifact(t, artifactRepo)
	// No file stored in downloader → download will fail

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	// Verify failure audit event
	events := auditRepo.GetEvents()
	foundFailure := false
	for _, e := range events {
		if e.Action == "inbody.extraction_failed" {
			foundFailure = true
		}
	}
	if !foundFailure {
		t.Error("expected audit event 'inbody.extraction_failed'")
	}
}

func TestExtractInBody_PDFExtractionFails_ReturnsError(t *testing.T) {
	uc, artifactRepo, fileDownloader, pdfExtractor, _, _ := newExtractInBodyTestDeps()
	artifact := createTestArtifact(t, artifactRepo)

	fileDownloader.PutFile(artifact.StorageKey, []byte("corrupt-pdf"))
	pdfExtractor.ForceError = fmt.Errorf("corrupt PDF: cannot parse")

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExtractInBody_NoMetricsInText_ReturnsError(t *testing.T) {
	uc, artifactRepo, fileDownloader, pdfExtractor, _, _ := newExtractInBodyTestDeps()
	artifact := createTestArtifact(t, artifactRepo)

	pdfContent := "no-metrics-pdf"
	fileDownloader.PutFile(artifact.StorageKey, []byte(pdfContent))
	pdfExtractor.SetTextForContent(pdfContent, "This is not an InBody report.")

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExtractInBody_MeasuredAtUsesCurrentTime(t *testing.T) {
	uc, artifactRepo, fileDownloader, pdfExtractor, measurementRepo, _ := newExtractInBodyTestDeps()
	artifact := createTestArtifact(t, artifactRepo)

	pdfContent := "with-time-pdf"
	fileDownloader.PutFile(artifact.StorageKey, []byte(pdfContent))
	pdfExtractor.SetTextForContent(pdfContent, sampleInBodyText)

	before := time.Now()
	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	after := time.Now()

	measurements := measurementRepo.GetAll()
	for _, m := range measurements {
		if m.MeasuredAt.Before(before) || m.MeasuredAt.After(after) {
			t.Errorf("measured_at %v not between %v and %v", m.MeasuredAt, before, after)
		}
	}
}

func TestExtractInBody_RecordedByIsCoachID(t *testing.T) {
	uc, artifactRepo, fileDownloader, pdfExtractor, measurementRepo, _ := newExtractInBodyTestDeps()
	artifact := createTestArtifact(t, artifactRepo)

	pdfContent := "recorded-by-pdf"
	fileDownloader.PutFile(artifact.StorageKey, []byte(pdfContent))
	pdfExtractor.SetTextForContent(pdfContent, sampleInBodyText)

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	measurements := measurementRepo.GetAll()
	for _, m := range measurements {
		if m.RecordedBy != "coach-1" {
			t.Errorf("expected recorded_by 'coach-1', got %q", m.RecordedBy)
		}
	}
}

func TestExtractInBody_FindByArtifactID_ReturnsMeasurements(t *testing.T) {
	uc, artifactRepo, fileDownloader, pdfExtractor, measurementRepo, _ := newExtractInBodyTestDeps()
	artifact := createTestArtifact(t, artifactRepo)

	pdfContent := "find-by-artifact-pdf"
	fileDownloader.PutFile(artifact.StorageKey, []byte(pdfContent))
	pdfExtractor.SetTextForContent(pdfContent, sampleInBodyText)

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, err := measurementRepo.FindByArtifactID(context.Background(), artifact.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(found) != 6 {
		t.Errorf("expected 6 measurements for artifact, got %d", len(found))
	}
}
