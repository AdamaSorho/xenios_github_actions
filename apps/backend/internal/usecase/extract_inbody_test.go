package usecase

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

// plainTextExtractor returns raw bytes as text (for testing with simulated InBody text).
func plainTextExtractor(pdfBytes []byte) (string, error) {
	return string(pdfBytes), nil
}

func newExtractInBodyUseCase() (
	*ExtractInBodyUseCase,
	*repository.InMemoryArtifactRepository,
	*repository.InMemoryMeasurementRepository,
	*repository.InMemoryFileDownloader,
	*repository.InMemoryAuditRepository,
) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	fileDownloader := repository.NewInMemoryFileDownloader()
	auditRepo := repository.NewInMemoryAuditRepository()
	extractor := repository.NewInBodyTextExtractor(plainTextExtractor)

	uc := NewExtractInBodyUseCase(artifactRepo, measurementRepo, fileDownloader, extractor, auditRepo)
	return uc, artifactRepo, measurementRepo, fileDownloader, auditRepo
}

func createUploadedInBodyArtifact(t *testing.T, repo *repository.InMemoryArtifactRepository) *entities.Artifact {
	t.Helper()
	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "inbody-scan.pdf",
		FileType:    "application/pdf",
		FileSize:    2048,
		StorageKey:  "client-1/document/test-id.pdf",
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

func makeJob(t *testing.T, artifactID string) *entities.Job {
	t.Helper()
	payload, err := json.Marshal(ExtractInBodyPayload{ArtifactID: artifactID})
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	return &entities.Job{
		ID:          "job-1",
		Type:        entities.JobTypeExtractInBody,
		Payload:     payload,
		Status:      entities.JobStatusActive,
		Attempt:     1,
		MaxAttempts: 3,
	}
}

const sampleInBodyText = `
InBody 570 Body Composition Analysis

Body Composition Analysis
Weight: 85.4 kg
Skeletal Muscle Mass: 35.2 kg
Body Fat Percentage: 22.3 %
Total Body Water: 42.1 L
Lean Body Mass: 66.4 kg
BMR: 1847 kcal
`

func TestExtractInBody_HappyPath_StoresAllMeasurements(t *testing.T) {
	uc, artifactRepo, measurementRepo, fileDownloader, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileDownloader.PutFile(artifact.StorageKey, []byte(sampleInBodyText))
	job := makeJob(t, artifact.ID)

	output, err := uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.FieldCount != 6 {
		t.Errorf("expected 6 fields, got %d", output.FieldCount)
	}
	if output.IsPartial {
		t.Error("expected full extraction, not partial")
	}
	if len(output.Measurements) != 6 {
		t.Errorf("expected 6 measurements, got %d", len(output.Measurements))
	}

	// Verify measurements in storage
	stored := measurementRepo.GetAll()
	if len(stored) != 6 {
		t.Errorf("expected 6 stored measurements, got %d", len(stored))
	}
}

func TestExtractInBody_HappyPath_MeasurementsHaveCorrectArtifactID(t *testing.T) {
	uc, artifactRepo, _, fileDownloader, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileDownloader.PutFile(artifact.StorageKey, []byte(sampleInBodyText))
	job := makeJob(t, artifact.ID)

	output, err := uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, m := range output.Measurements {
		if m.ArtifactID != artifact.ID {
			t.Errorf("expected artifact_id %q, got %q", artifact.ID, m.ArtifactID)
		}
		if m.ClientID != artifact.ClientID {
			t.Errorf("expected client_id %q, got %q", artifact.ClientID, m.ClientID)
		}
	}
}

func TestExtractInBody_HappyPath_MeasurementsHaveCorrectTypes(t *testing.T) {
	uc, artifactRepo, _, fileDownloader, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileDownloader.PutFile(artifact.StorageKey, []byte(sampleInBodyText))
	job := makeJob(t, artifact.ID)

	output, err := uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	typeSet := make(map[entities.MeasurementType]bool)
	for _, m := range output.Measurements {
		typeSet[m.MeasurementType] = true
	}

	expected := []entities.MeasurementType{
		entities.MeasurementTypeWeight,
		entities.MeasurementTypeBodyFatPct,
		entities.MeasurementTypeSkeletalMuscleMass,
		entities.MeasurementTypeBMR,
		entities.MeasurementTypeTotalBodyWater,
		entities.MeasurementTypeLeanBodyMass,
	}
	for _, et := range expected {
		if !typeSet[et] {
			t.Errorf("expected measurement type %q to be present", et)
		}
	}
}

func TestExtractInBody_PartialExtraction_FlaggedCorrectly(t *testing.T) {
	partialText := `
Weight: 85.4 kg
Body Fat Percentage: 22.3
SMM: 35.2 kg
`
	uc, artifactRepo, _, fileDownloader, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileDownloader.PutFile(artifact.StorageKey, []byte(partialText))
	job := makeJob(t, artifact.ID)

	output, err := uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.IsPartial {
		t.Error("expected partial extraction")
	}
	if output.FieldCount != 3 {
		t.Errorf("expected 3 fields, got %d", output.FieldCount)
	}
	// Each measurement should be flagged as partial
	for _, m := range output.Measurements {
		if !m.PartialExtraction {
			t.Errorf("expected partial_extraction flag on measurement %s", m.MeasurementType)
		}
	}
}

func TestExtractInBody_EmptyExtraction_ReturnsError(t *testing.T) {
	unrecognizedText := `This is not an InBody report at all.`
	uc, artifactRepo, _, fileDownloader, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileDownloader.PutFile(artifact.StorageKey, []byte(unrecognizedText))
	job := makeJob(t, artifact.ID)

	_, err := uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for empty extraction")
	}
}

func TestExtractInBody_InvalidPayload_ReturnsError(t *testing.T) {
	uc, _, _, _, _ := newExtractInBodyUseCase()

	job := &entities.Job{
		ID:      "job-1",
		Type:    entities.JobTypeExtractInBody,
		Payload: json.RawMessage(`{invalid json}`),
	}

	_, err := uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for invalid payload")
	}
}

func TestExtractInBody_MissingArtifactID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newExtractInBodyUseCase()

	payload, _ := json.Marshal(ExtractInBodyPayload{ArtifactID: ""})
	job := &entities.Job{
		ID:      "job-1",
		Type:    entities.JobTypeExtractInBody,
		Payload: payload,
	}

	_, err := uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for missing artifact_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractInBody_ArtifactNotFound_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newExtractInBodyUseCase()
	job := makeJob(t, "nonexistent")

	_, err := uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for missing artifact")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractInBody_ArtifactNotUploaded_ReturnsValidationError(t *testing.T) {
	uc, artifactRepo, _, _, _ := newExtractInBodyUseCase()

	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "scan.pdf",
		FileType:    "application/pdf",
		FileSize:    1024,
		StorageKey:  "client-1/document/test.pdf",
		Type:        entities.ArtifactTypeDocument,
		Status:      entities.ArtifactStatusPending,
		ContentType: "application/pdf",
	}
	created, _ := artifactRepo.Create(context.Background(), art)
	job := makeJob(t, created.ID)

	_, err := uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for non-uploaded artifact")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractInBody_NonDocumentArtifact_ReturnsValidationError(t *testing.T) {
	uc, artifactRepo, _, _, _ := newExtractInBodyUseCase()

	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "photo.jpg",
		FileType:    "image/jpeg",
		FileSize:    1024,
		StorageKey:  "client-1/image/test.jpg",
		Type:        entities.ArtifactTypeImage,
		Status:      entities.ArtifactStatusUploaded,
		ContentType: "image/jpeg",
	}
	created, _ := artifactRepo.Create(context.Background(), art)
	job := makeJob(t, created.ID)

	_, err := uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for non-document artifact")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractInBody_DownloadFailure_ReturnsError(t *testing.T) {
	uc, artifactRepo, _, _, _ := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	// Don't put the file in the downloader
	job := makeJob(t, artifact.ID)

	_, err := uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error when file download fails")
	}
}

func TestExtractInBody_AuditEventLogged_OnSuccess(t *testing.T) {
	uc, artifactRepo, _, fileDownloader, auditRepo := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileDownloader.PutFile(artifact.StorageKey, []byte(sampleInBodyText))
	job := makeJob(t, artifact.ID)

	_, err := uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	if len(events) == 0 {
		t.Fatal("expected audit event to be logged")
	}
	found := false
	for _, e := range events {
		if e.Action == "extraction.inbody_success" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'extraction.inbody_success' audit event")
	}
}

func TestExtractInBody_AuditEventLogged_OnPartialExtraction(t *testing.T) {
	partialText := `Weight: 85.4 kg`
	uc, artifactRepo, _, fileDownloader, auditRepo := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileDownloader.PutFile(artifact.StorageKey, []byte(partialText))
	job := makeJob(t, artifact.ID)

	_, err := uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	found := false
	for _, e := range events {
		if e.Action == "extraction.inbody_partial" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'extraction.inbody_partial' audit event")
	}
}

func TestExtractInBody_AuditEventLogged_OnFailure(t *testing.T) {
	unrecognizedText := `Not an InBody report`
	uc, artifactRepo, _, fileDownloader, auditRepo := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	fileDownloader.PutFile(artifact.StorageKey, []byte(unrecognizedText))
	job := makeJob(t, artifact.ID)

	_, _ = uc.Execute(context.Background(), job)

	events := auditRepo.GetEvents()
	found := false
	for _, e := range events {
		if e.Action == "extraction.inbody_failed" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'extraction.inbody_failed' audit event")
	}
}

func TestExtractInBody_AuditEventLogged_OnDownloadFailure(t *testing.T) {
	uc, artifactRepo, _, _, auditRepo := newExtractInBodyUseCase()
	artifact := createUploadedInBodyArtifact(t, artifactRepo)
	// Don't put file in downloader to trigger download failure
	job := makeJob(t, artifact.ID)

	_, _ = uc.Execute(context.Background(), job)

	events := auditRepo.GetEvents()
	found := false
	for _, e := range events {
		if e.Action == "extraction.inbody_failed" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'extraction.inbody_failed' audit event for download failure")
	}
}
