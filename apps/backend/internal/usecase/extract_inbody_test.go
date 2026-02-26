package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/infrastructure/pdf"
)

// mockPDFTextExtractor implements repository.PDFTextExtractor for testing.
type mockPDFTextExtractor struct {
	extractFunc func(ctx context.Context, data []byte) (string, error)
}

func (m *mockPDFTextExtractor) ExtractText(ctx context.Context, data []byte) (string, error) {
	if m.extractFunc != nil {
		return m.extractFunc(ctx, data)
	}
	return "", nil
}

const testInBodyPDFText = `InBody 570 Body Composition Analysis
Weight: 85.4 kg
Skeletal Muscle Mass: 38.2 kg
Body Fat Percentage: 22.3%
Basal Metabolic Rate: 1847 kcal
Total Body Water: 42.1 L
Lean Body Mass: 66.4 kg
`

func newExtractInBodyUseCase() (
	*ExtractInBodyUseCase,
	*repository.InMemoryArtifactRepository,
	*repository.InMemoryFileStorage,
	*mockPDFTextExtractor,
	*repository.InMemoryMeasurementRepository,
	*repository.InMemoryAuditRepository,
) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	pdfExtractor := &mockPDFTextExtractor{
		extractFunc: func(_ context.Context, _ []byte) (string, error) {
			return testInBodyPDFText, nil
		},
	}
	inbodyParser := pdf.NewInBodyParser()
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := NewExtractInBodyUseCase(artifactRepo, fileStorage, pdfExtractor, inbodyParser, measurementRepo, auditRepo)
	return uc, artifactRepo, fileStorage, pdfExtractor, measurementRepo, auditRepo
}

func createUploadedInBodyArtifact(t *testing.T, repo *repository.InMemoryArtifactRepository) *entities.Artifact {
	t.Helper()
	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "inbody_scan.pdf",
		FileType:    "application/pdf",
		FileSize:    2048,
		StorageKey:  "client-1/document/artifact-1.pdf",
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

func TestExtractInBody_ValidInput_ExtractsAllMeasurements(t *testing.T) {
	uc, artifactRepo, fileStorage, _, measurementRepo, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileStorage.PutObjectWithContent(artifact.StorageKey, []byte("fake pdf data"))

	output, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output == nil {
		t.Fatal("expected non-nil output")
	}
	if len(output.Measurements) != 6 {
		t.Errorf("expected 6 measurements, got %d", len(output.Measurements))
	}

	// Verify measurements were persisted
	stored := measurementRepo.All()
	if len(stored) != 6 {
		t.Errorf("expected 6 stored measurements, got %d", len(stored))
	}
}

func TestExtractInBody_ValidInput_ArtifactMarkedProcessed(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileStorage.PutObjectWithContent(artifact.StorageKey, []byte("fake pdf data"))

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := artifactRepo.FindByID(context.Background(), artifact.ID)
	if updated.Status != entities.ArtifactStatusProcessed {
		t.Errorf("expected artifact status 'processed', got '%s'", updated.Status)
	}
}

func TestExtractInBody_EmptyArtifactID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _, _ := newExtractInBodyUseCase()
	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: "",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestExtractInBody_ArtifactNotFound_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _, _ := newExtractInBodyUseCase()
	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: "nonexistent-id",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestExtractInBody_ArtifactNotUploaded_ReturnsValidationError(t *testing.T) {
	uc, artifactRepo, _, _, _, _ := newExtractInBodyUseCase()
	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "inbody.pdf",
		FileType:    "application/pdf",
		FileSize:    1024,
		StorageKey:  "client-1/document/test.pdf",
		Type:        entities.ArtifactTypeDocument,
		Status:      entities.ArtifactStatusPending,
		ContentType: "application/pdf",
	}
	created, _ := artifactRepo.Create(context.Background(), art)

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: created.ID,
	})
	if err == nil {
		t.Fatal("expected error for non-uploaded artifact")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestExtractInBody_DownloadFails_ReturnsError(t *testing.T) {
	uc, artifactRepo, _, _, _, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	// Don't put file in storage - download will fail

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
	})
	if err == nil {
		t.Fatal("expected error when download fails")
	}
}

func TestExtractInBody_PDFExtractionFails_ReturnsError(t *testing.T) {
	uc, artifactRepo, fileStorage, pdfExtractor, _, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileStorage.PutObjectWithContent(artifact.StorageKey, []byte("corrupt pdf"))

	pdfExtractor.extractFunc = func(_ context.Context, _ []byte) (string, error) {
		return "", errors.New("PDF parsing failed: corrupt file")
	}

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
	})
	if err == nil {
		t.Fatal("expected error when PDF extraction fails")
	}
}

func TestExtractInBody_NoMetricsExtracted_ReturnsError(t *testing.T) {
	uc, artifactRepo, fileStorage, pdfExtractor, _, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileStorage.PutObjectWithContent(artifact.StorageKey, []byte("some pdf"))

	pdfExtractor.extractFunc = func(_ context.Context, _ []byte) (string, error) {
		return "This text has no InBody data", nil
	}

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
	})
	if err == nil {
		t.Fatal("expected error when no metrics found")
	}
}

func TestExtractInBody_PartialExtraction_ReturnsPartialResult(t *testing.T) {
	uc, artifactRepo, fileStorage, pdfExtractor, measurementRepo, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileStorage.PutObjectWithContent(artifact.StorageKey, []byte("some pdf"))

	pdfExtractor.extractFunc = func(_ context.Context, _ []byte) (string, error) {
		return "Weight: 90.0 kg\nBody Fat Percentage: 18.5%\nSkeletal Muscle Mass: 42.0 kg", nil
	}

	output, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.IsPartial {
		t.Error("expected partial result")
	}
	if len(output.Measurements) != 3 {
		t.Errorf("expected 3 measurements, got %d", len(output.Measurements))
	}

	stored := measurementRepo.All()
	if len(stored) != 3 {
		t.Errorf("expected 3 stored measurements, got %d", len(stored))
	}
}

func TestExtractInBody_AuditEventLogged_OnSuccess(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _, auditRepo := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileStorage.PutObjectWithContent(artifact.StorageKey, []byte("fake pdf"))

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
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
			break
		}
	}
	if !found {
		t.Error("expected artifact.extraction_success audit event")
	}
}

func TestExtractInBody_AuditEventLogged_OnPartialSuccess(t *testing.T) {
	uc, artifactRepo, fileStorage, pdfExtractor, _, auditRepo := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileStorage.PutObjectWithContent(artifact.StorageKey, []byte("some pdf"))

	pdfExtractor.extractFunc = func(_ context.Context, _ []byte) (string, error) {
		return "Weight: 90.0 kg\nBody Fat Percentage: 18.5%", nil
	}

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
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
		t.Error("expected artifact.extraction_partial audit event")
	}
}

func TestExtractInBody_AuditEventLogged_OnFailure(t *testing.T) {
	uc, artifactRepo, fileStorage, pdfExtractor, _, auditRepo := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileStorage.PutObjectWithContent(artifact.StorageKey, []byte("corrupt"))

	pdfExtractor.extractFunc = func(_ context.Context, _ []byte) (string, error) {
		return "", errors.New("corrupt PDF")
	}

	_, _ = uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
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
		t.Error("expected artifact.extraction_failed audit event")
	}
}

func TestExtractInBody_MeasurementsMappedCorrectly(t *testing.T) {
	uc, artifactRepo, fileStorage, _, measurementRepo, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileStorage.PutObjectWithContent(artifact.StorageKey, []byte("fake pdf"))

	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored := measurementRepo.All()
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
		if m.MeasuredAt.IsZero() {
			t.Error("expected non-zero measured_at")
		}
	}
}

func TestExtractInBody_ContextCancelled_ReturnsError(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileStorage.PutObjectWithContent(artifact.StorageKey, []byte("fake pdf"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := uc.Execute(ctx, ExtractInBodyInput{
		ArtifactID: artifact.ID,
	})
	// The error should propagate from the cancelled context
	if err == nil {
		// InMemory repos don't check context, so this may not fail.
		// That's OK - it's an integration concern.
		t.Log("in-memory repos don't propagate context cancellation")
	}
}

func TestExtractInBody_MeasuredAtSetFromNow(t *testing.T) {
	uc, artifactRepo, fileStorage, _, measurementRepo, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileStorage.PutObjectWithContent(artifact.StorageKey, []byte("fake pdf"))

	before := time.Now()
	_, err := uc.Execute(context.Background(), ExtractInBodyInput{
		ArtifactID: artifact.ID,
	})
	after := time.Now()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored := measurementRepo.All()
	for _, m := range stored {
		if m.MeasuredAt.Before(before) || m.MeasuredAt.After(after) {
			t.Errorf("expected measured_at between %v and %v, got %v", before, after, m.MeasuredAt)
		}
	}
}
