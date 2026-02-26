package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
	domainrepo "github.com/xenios/backend/internal/domain/repository"
)

// mockPDFExtractor implements repository.PDFExtractor for testing.
type mockPDFExtractor struct {
	extractFn func(ctx context.Context, pdfData []byte) (*entities.InBodyResult, error)
}

func (m *mockPDFExtractor) ExtractInBody(ctx context.Context, pdfData []byte) (*entities.InBodyResult, error) {
	if m.extractFn != nil {
		return m.extractFn(ctx, pdfData)
	}
	return nil, fmt.Errorf("not implemented")
}

var _ domainrepo.PDFExtractor = &mockPDFExtractor{}

func fullInBodyResult() *entities.InBodyResult {
	w := 85.5
	bf := 22.3
	smm := 35.2
	bmr := 1847.0
	tbw := 42.1
	lbm := 66.5
	return &entities.InBodyResult{
		Weight:             &w,
		WeightUnit:         "kg",
		BodyFatPct:         &bf,
		SkeletalMuscleMass: &smm,
		SkeletalMuscleUnit: "kg",
		BMR:                &bmr,
		TotalBodyWater:     &tbw,
		LeanBodyMass:       &lbm,
		LeanBodyMassUnit:   "kg",
		MeasuredAt:         time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		IsPartial:          false,
	}
}

func partialInBodyResult() *entities.InBodyResult {
	w := 85.5
	bf := 22.3
	return &entities.InBodyResult{
		Weight:     &w,
		WeightUnit: "kg",
		BodyFatPct: &bf,
		MeasuredAt: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		IsPartial:  true,
	}
}

type extractInBodyDeps struct {
	artifactRepo    *repository.InMemoryArtifactRepository
	measurementRepo *repository.InMemoryMeasurementRepository
	fileStorage     *repository.InMemoryFileStorage
	pdfExtractor    *mockPDFExtractor
	auditRepo       *repository.InMemoryAuditRepository
	uc              *ExtractInBodyUseCase
}

func setupExtractInBody() *extractInBodyDeps {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	pdfExtractor := &mockPDFExtractor{}
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := NewExtractInBodyUseCase(artifactRepo, measurementRepo, fileStorage, pdfExtractor, auditRepo)
	return &extractInBodyDeps{
		artifactRepo:    artifactRepo,
		measurementRepo: measurementRepo,
		fileStorage:     fileStorage,
		pdfExtractor:    pdfExtractor,
		auditRepo:       auditRepo,
		uc:              uc,
	}
}

func makeJob(t *testing.T, artifactID, clientID, coachID string) *entities.Job {
	t.Helper()
	payload, err := json.Marshal(ExtractInBodyPayload{
		ArtifactID: artifactID,
		ClientID:   clientID,
		CoachID:    coachID,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return &entities.Job{
		ID:          "job-1",
		Type:        entities.JobTypeExtractInBody,
		Payload:     payload,
		Status:      entities.JobStatusActive,
		Attempt:     1,
		MaxAttempts: entities.MaxRetryAttempts,
		CreatedAt:   time.Now(),
	}
}

func createUploadedInBodyArtifact(t *testing.T, repo *repository.InMemoryArtifactRepository) *entities.Artifact {
	t.Helper()
	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "inbody_scan.pdf",
		FileType:    "application/pdf",
		FileSize:    2048,
		StorageKey:  "client-1/document/art-1.pdf",
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

func TestExtractInBody_HappyPath_StoresMeasurements(t *testing.T) {
	deps := setupExtractInBody()
	artifact := createUploadedInBodyArtifact(t, deps.artifactRepo)
	deps.fileStorage.PutObjectWithData(artifact.StorageKey, []byte("fake-pdf-data"))
	deps.pdfExtractor.extractFn = func(_ context.Context, _ []byte) (*entities.InBodyResult, error) {
		return fullInBodyResult(), nil
	}

	job := makeJob(t, artifact.ID, "client-1", "coach-1")
	err := deps.uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify measurements were stored
	measurements := deps.measurementRepo.GetAll()
	if len(measurements) != 6 {
		t.Fatalf("expected 6 measurements, got %d", len(measurements))
	}

	// Verify each measurement has the correct artifact_id
	for _, m := range measurements {
		if m.ArtifactID != artifact.ID {
			t.Errorf("expected artifact_id %q, got %q", artifact.ID, m.ArtifactID)
		}
		if m.ClientID != "client-1" {
			t.Errorf("expected client_id 'client-1', got %q", m.ClientID)
		}
	}
}

func TestExtractInBody_HappyPath_LogsSuccessAuditEvent(t *testing.T) {
	deps := setupExtractInBody()
	artifact := createUploadedInBodyArtifact(t, deps.artifactRepo)
	deps.fileStorage.PutObjectWithData(artifact.StorageKey, []byte("fake-pdf-data"))
	deps.pdfExtractor.extractFn = func(_ context.Context, _ []byte) (*entities.InBodyResult, error) {
		return fullInBodyResult(), nil
	}

	job := makeJob(t, artifact.ID, "client-1", "coach-1")
	err := deps.uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := deps.auditRepo.GetEvents()
	if len(events) == 0 {
		t.Fatal("expected audit event to be logged")
	}
	lastEvent := events[len(events)-1]
	if lastEvent.Action != "artifact.extraction_success" {
		t.Errorf("expected action 'artifact.extraction_success', got %q", lastEvent.Action)
	}
	if lastEvent.EntityID != artifact.ID {
		t.Errorf("expected entity_id %q, got %q", artifact.ID, lastEvent.EntityID)
	}
}

func TestExtractInBody_PartialExtraction_LogsPartialAuditEvent(t *testing.T) {
	deps := setupExtractInBody()
	artifact := createUploadedInBodyArtifact(t, deps.artifactRepo)
	deps.fileStorage.PutObjectWithData(artifact.StorageKey, []byte("fake-pdf-data"))
	deps.pdfExtractor.extractFn = func(_ context.Context, _ []byte) (*entities.InBodyResult, error) {
		return partialInBodyResult(), nil
	}

	job := makeJob(t, artifact.ID, "client-1", "coach-1")
	err := deps.uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only 2 measurements should be stored
	measurements := deps.measurementRepo.GetAll()
	if len(measurements) != 2 {
		t.Fatalf("expected 2 measurements, got %d", len(measurements))
	}

	events := deps.auditRepo.GetEvents()
	if len(events) == 0 {
		t.Fatal("expected audit event to be logged")
	}
	lastEvent := events[len(events)-1]
	if lastEvent.Action != "artifact.extraction_partial" {
		t.Errorf("expected action 'artifact.extraction_partial', got %q", lastEvent.Action)
	}
}

func TestExtractInBody_NilJob_ReturnsValidationError(t *testing.T) {
	deps := setupExtractInBody()
	err := deps.uc.Execute(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractInBody_InvalidPayload_ReturnsError(t *testing.T) {
	deps := setupExtractInBody()
	job := &entities.Job{
		ID:      "job-1",
		Type:    entities.JobTypeExtractInBody,
		Payload: []byte(`{invalid json`),
	}
	err := deps.uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for invalid payload")
	}
}

func TestExtractInBody_EmptyArtifactID_ReturnsValidationError(t *testing.T) {
	deps := setupExtractInBody()
	job := makeJob(t, "", "client-1", "coach-1")
	err := deps.uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractInBody_EmptyClientID_ReturnsValidationError(t *testing.T) {
	deps := setupExtractInBody()
	job := makeJob(t, "art-1", "", "coach-1")
	err := deps.uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractInBody_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	deps := setupExtractInBody()
	job := makeJob(t, "art-1", "client-1", "")
	err := deps.uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractInBody_ArtifactNotFound_ReturnsValidationError(t *testing.T) {
	deps := setupExtractInBody()
	job := makeJob(t, "nonexistent", "client-1", "coach-1")
	err := deps.uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractInBody_ArtifactNotUploaded_ReturnsValidationError(t *testing.T) {
	deps := setupExtractInBody()
	// Create a pending (not uploaded) artifact
	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "scan.pdf",
		FileType:    "application/pdf",
		FileSize:    1024,
		StorageKey:  "client-1/document/art-2.pdf",
		Type:        entities.ArtifactTypeDocument,
		Status:      entities.ArtifactStatusPending,
		ContentType: "application/pdf",
	}
	created, _ := deps.artifactRepo.Create(context.Background(), art)

	job := makeJob(t, created.ID, "client-1", "coach-1")
	err := deps.uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractInBody_DownloadFails_LogsFailedAudit(t *testing.T) {
	deps := setupExtractInBody()
	artifact := createUploadedInBodyArtifact(t, deps.artifactRepo)
	// Don't put any file data in storage

	job := makeJob(t, artifact.ID, "client-1", "coach-1")
	err := deps.uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error when download fails")
	}

	events := deps.auditRepo.GetEvents()
	if len(events) == 0 {
		t.Fatal("expected failure audit event")
	}
	if events[0].Action != "artifact.extraction_failed" {
		t.Errorf("expected action 'artifact.extraction_failed', got %q", events[0].Action)
	}
}

func TestExtractInBody_ExtractionFails_LogsFailedAudit(t *testing.T) {
	deps := setupExtractInBody()
	artifact := createUploadedInBodyArtifact(t, deps.artifactRepo)
	deps.fileStorage.PutObjectWithData(artifact.StorageKey, []byte("corrupt-pdf"))
	deps.pdfExtractor.extractFn = func(_ context.Context, _ []byte) (*entities.InBodyResult, error) {
		return nil, fmt.Errorf("cannot parse PDF")
	}

	job := makeJob(t, artifact.ID, "client-1", "coach-1")
	err := deps.uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error when extraction fails")
	}

	events := deps.auditRepo.GetEvents()
	if len(events) == 0 {
		t.Fatal("expected failure audit event")
	}
	if events[0].Action != "artifact.extraction_failed" {
		t.Errorf("expected action 'artifact.extraction_failed', got %q", events[0].Action)
	}
}

func TestExtractInBody_StoreMeasurementsFails_LogsFailedAudit(t *testing.T) {
	// Use a custom measurement repo that fails
	deps := setupExtractInBody()
	artifact := createUploadedInBodyArtifact(t, deps.artifactRepo)
	deps.fileStorage.PutObjectWithData(artifact.StorageKey, []byte("fake-pdf-data"))
	deps.pdfExtractor.extractFn = func(_ context.Context, _ []byte) (*entities.InBodyResult, error) {
		return fullInBodyResult(), nil
	}

	// Replace measurement repo with a failing one
	failingMeasRepo := &mockMeasurementRepo{
		createBatchFn: func(_ context.Context, _ []*entities.Measurement) ([]*entities.Measurement, error) {
			return nil, fmt.Errorf("database error")
		},
	}
	deps.uc.measurementRepo = failingMeasRepo

	job := makeJob(t, artifact.ID, "client-1", "coach-1")
	err := deps.uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error when store fails")
	}

	events := deps.auditRepo.GetEvents()
	if len(events) == 0 {
		t.Fatal("expected failure audit event")
	}
	if events[0].Action != "artifact.extraction_failed" {
		t.Errorf("expected action 'artifact.extraction_failed', got %q", events[0].Action)
	}
}

func TestExtractInBody_MeasurementsByArtifactID_ReturnsCorrect(t *testing.T) {
	deps := setupExtractInBody()
	artifact := createUploadedInBodyArtifact(t, deps.artifactRepo)
	deps.fileStorage.PutObjectWithData(artifact.StorageKey, []byte("fake-pdf-data"))
	deps.pdfExtractor.extractFn = func(_ context.Context, _ []byte) (*entities.InBodyResult, error) {
		return fullInBodyResult(), nil
	}

	job := makeJob(t, artifact.ID, "client-1", "coach-1")
	err := deps.uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Query measurements by artifact ID
	found, err := deps.measurementRepo.FindByArtifactID(context.Background(), artifact.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(found) != 6 {
		t.Fatalf("expected 6 measurements for artifact, got %d", len(found))
	}
}

// mockMeasurementRepo implements repository.MeasurementRepository for testing failures.
type mockMeasurementRepo struct {
	createBatchFn    func(ctx context.Context, measurements []*entities.Measurement) ([]*entities.Measurement, error)
	findByArtifactFn func(ctx context.Context, artifactID string) ([]*entities.Measurement, error)
}

func (m *mockMeasurementRepo) CreateBatch(ctx context.Context, measurements []*entities.Measurement) ([]*entities.Measurement, error) {
	if m.createBatchFn != nil {
		return m.createBatchFn(ctx, measurements)
	}
	return nil, nil
}

func (m *mockMeasurementRepo) FindByArtifactID(ctx context.Context, artifactID string) ([]*entities.Measurement, error) {
	if m.findByArtifactFn != nil {
		return m.findByArtifactFn(ctx, artifactID)
	}
	return nil, nil
}

var _ domainrepo.MeasurementRepository = &mockMeasurementRepo{}
