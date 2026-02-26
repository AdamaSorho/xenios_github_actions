package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newConfirmUploadUseCase() (*ConfirmUploadUseCase, *repository.InMemoryArtifactRepository, *repository.InMemoryFileStorage, *repository.InMemoryAuditRepository, *repository.InMemoryJobQueue) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	auditRepo := repository.NewInMemoryAuditRepository()
	jobQueue := repository.NewInMemoryJobQueue()

	uc := NewConfirmUploadUseCase(artifactRepo, fileStorage, auditRepo, jobQueue)
	return uc, artifactRepo, fileStorage, auditRepo, jobQueue
}

func createTestArtifact(t *testing.T, repo *repository.InMemoryArtifactRepository, fileName, contentType string, hint entities.DocumentSubtype) *entities.Artifact {
	t.Helper()
	art := &entities.Artifact{
		ClientID:        "client-1",
		CoachID:         "coach-1",
		FileName:        fileName,
		FileType:        contentType,
		FileSize:        1024,
		StorageKey:      "client-1/document/test-id.pdf",
		Type:            entities.ArtifactTypeDocument,
		Status:          entities.ArtifactStatusPending,
		ContentType:     contentType,
		DocumentSubtype: hint,
	}
	created, err := repo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("failed to create test artifact: %v", err)
	}
	return created
}

func createPendingArtifact(t *testing.T, repo *repository.InMemoryArtifactRepository) *entities.Artifact {
	t.Helper()
	return createTestArtifact(t, repo, "report.pdf", "application/pdf", "")
}

func createPendingArtifactWithHint(t *testing.T, repo *repository.InMemoryArtifactRepository, fileName, contentType string, hint entities.DocumentSubtype) *entities.Artifact {
	t.Helper()
	return createTestArtifact(t, repo, fileName, contentType, hint)
}

func TestConfirmUpload_ValidInput_UpdatesStatusToUploaded(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifact(t, artifactRepo)

	// Simulate file upload to storage
	fileStorage.PutObject(artifact.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Artifact.Status != entities.ArtifactStatusUploaded {
		t.Errorf("expected status 'uploaded', got '%s'", out.Artifact.Status)
	}
}

func TestConfirmUpload_EmptyArtifactID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newConfirmUploadUseCase()

	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
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

func TestConfirmUpload_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newConfirmUploadUseCase()

	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
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

func TestConfirmUpload_ArtifactNotFound_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newConfirmUploadUseCase()

	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
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

func TestConfirmUpload_WrongCoach_ReturnsAuthError(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifact(t, artifactRepo)
	fileStorage.PutObject(artifact.StorageKey)

	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
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

func TestConfirmUpload_AlreadyUploaded_ReturnsValidationError(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifact(t, artifactRepo)
	fileStorage.PutObject(artifact.StorageKey)

	// First confirmation
	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second confirmation should fail
	_, err = uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error for already uploaded artifact")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestConfirmUpload_FileNotInStorage_ReturnsValidationError(t *testing.T) {
	uc, artifactRepo, _, _, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifact(t, artifactRepo)

	// Don't put the file in storage
	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error when file not in storage")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestConfirmUpload_FileNotInStorage_MarksArtifactAsFailed(t *testing.T) {
	uc, artifactRepo, _, _, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifact(t, artifactRepo)

	// Don't put the file in storage
	_, _ = uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})

	// Check that artifact was marked as failed
	updated, err := artifactRepo.FindByID(context.Background(), artifact.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != entities.ArtifactStatusFailed {
		t.Errorf("expected status 'failed', got '%s'", updated.Status)
	}
}

func TestConfirmUpload_AuditEventLogged(t *testing.T) {
	uc, artifactRepo, fileStorage, auditRepo, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifact(t, artifactRepo)
	fileStorage.PutObject(artifact.StorageKey)

	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
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
	if events[0].Action != "artifact.upload_confirmed" {
		t.Errorf("expected action 'artifact.upload_confirmed', got '%s'", events[0].Action)
	}
}

// --- Classification and job enqueue tests ---

func TestConfirmUpload_PDFWithInBodyHint_ClassifiesAsInBodyPDF(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifactWithHint(t, artifactRepo, "inbody_scan.pdf", "application/pdf", entities.DocumentSubtypeInBodyPDF)
	fileStorage.PutObject(artifact.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Artifact.DocumentSubtype != entities.DocumentSubtypeInBodyPDF {
		t.Errorf("expected document_subtype %q, got %q", entities.DocumentSubtypeInBodyPDF, out.Artifact.DocumentSubtype)
	}
}

func TestConfirmUpload_PDFNoHint_ClassifiesAsOther(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifact(t, artifactRepo)
	fileStorage.PutObject(artifact.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Artifact.DocumentSubtype != entities.DocumentSubtypeOther {
		t.Errorf("expected document_subtype %q, got %q", entities.DocumentSubtypeOther, out.Artifact.DocumentSubtype)
	}
}

func TestConfirmUpload_CSVWithLabHint_ClassifiesAsLabCSV(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifactWithHint(t, artifactRepo, "bloodwork.csv", "text/csv", entities.DocumentSubtypeLabCSV)
	fileStorage.PutObject(artifact.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Artifact.DocumentSubtype != entities.DocumentSubtypeLabCSV {
		t.Errorf("expected document_subtype %q, got %q", entities.DocumentSubtypeLabCSV, out.Artifact.DocumentSubtype)
	}
}

func TestConfirmUpload_AudioFile_ClassifiesAsAudio(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCase()
	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "session.mp3",
		FileType:    "audio/mpeg",
		FileSize:    5000000,
		StorageKey:  "client-1/audio/test-id.mp3",
		Type:        entities.ArtifactTypeAudio,
		Status:      entities.ArtifactStatusPending,
		ContentType: "audio/mpeg",
	}
	created, err := artifactRepo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}
	fileStorage.PutObject(created.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: created.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Artifact.DocumentSubtype != entities.DocumentSubtypeAudio {
		t.Errorf("expected document_subtype %q, got %q", entities.DocumentSubtypeAudio, out.Artifact.DocumentSubtype)
	}
}

func TestConfirmUpload_EnqueuesExtractionJob(t *testing.T) {
	uc, artifactRepo, fileStorage, _, jobQueue := newConfirmUploadUseCase()
	artifact := createPendingArtifactWithHint(t, artifactRepo, "inbody_scan.pdf", "application/pdf", entities.DocumentSubtypeInBodyPDF)
	fileStorage.PutObject(artifact.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.JobID == "" {
		t.Fatal("expected non-empty job_id")
	}

	jobs := jobQueue.GetJobs()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Type != entities.JobTypeExtractInBody {
		t.Errorf("expected job type %q, got %q", entities.JobTypeExtractInBody, jobs[0].Type)
	}
}

func TestConfirmUpload_OtherSubtype_EnqueuesClassifyDocumentJob(t *testing.T) {
	uc, artifactRepo, fileStorage, _, jobQueue := newConfirmUploadUseCase()
	artifact := createPendingArtifact(t, artifactRepo) // PDF without hint
	fileStorage.PutObject(artifact.StorageKey)

	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	jobs := jobQueue.GetJobs()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Type != entities.JobTypeClassifyDocument {
		t.Errorf("expected job type %q, got %q", entities.JobTypeClassifyDocument, jobs[0].Type)
	}
}

func TestConfirmUpload_AudioSubtype_EnqueuesTranscribeAudioJob(t *testing.T) {
	uc, artifactRepo, fileStorage, _, jobQueue := newConfirmUploadUseCase()
	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "session.wav",
		FileType:    "audio/wav",
		FileSize:    5000000,
		StorageKey:  "client-1/audio/test-wav.wav",
		Type:        entities.ArtifactTypeAudio,
		Status:      entities.ArtifactStatusPending,
		ContentType: "audio/wav",
	}
	created, err := artifactRepo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}
	fileStorage.PutObject(created.StorageKey)

	_, err = uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: created.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	jobs := jobQueue.GetJobs()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Type != entities.JobTypeTranscribeAudio {
		t.Errorf("expected job type %q, got %q", entities.JobTypeTranscribeAudio, jobs[0].Type)
	}
}

func TestConfirmUpload_ClassificationAuditEventLogged(t *testing.T) {
	uc, artifactRepo, fileStorage, auditRepo, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifactWithHint(t, artifactRepo, "inbody.pdf", "application/pdf", entities.DocumentSubtypeInBodyPDF)
	fileStorage.PutObject(artifact.StorageKey)

	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	if len(events) < 2 {
		t.Fatalf("expected at least 2 audit events, got %d", len(events))
	}

	// Second event should be classification
	classifyEvent := events[1]
	if classifyEvent.Action != "artifact.classified" {
		t.Errorf("expected action 'artifact.classified', got '%s'", classifyEvent.Action)
	}
	if classifyEvent.Metadata["document_subtype"] != "inbody_pdf" {
		t.Errorf("expected metadata document_subtype 'inbody_pdf', got '%v'", classifyEvent.Metadata["document_subtype"])
	}
}

func TestConfirmUpload_NilJobQueue_SucceedsWithoutEnqueue(t *testing.T) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := NewConfirmUploadUseCase(artifactRepo, fileStorage, auditRepo, nil)
	artifact := createPendingArtifact(t, artifactRepo)
	fileStorage.PutObject(artifact.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.JobID != "" {
		t.Errorf("expected empty job_id when no queue, got %q", out.JobID)
	}
}
