package usecase

import (
	"context"
	"encoding/json"
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

func createPendingArtifact(t *testing.T, repo *repository.InMemoryArtifactRepository) *entities.Artifact {
	t.Helper()
	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "report.pdf",
		FileType:    "application/pdf",
		FileSize:    1024,
		StorageKey:  "client-1/document/test-id.pdf",
		Type:        entities.ArtifactTypeDocument,
		Status:      entities.ArtifactStatusPending,
		ContentType: "application/pdf",
	}
	created, err := repo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("failed to create test artifact: %v", err)
	}
	return created
}

func createPendingArtifactWithFile(t *testing.T, repo *repository.InMemoryArtifactRepository, fileName, contentType string, artType entities.ArtifactType) *entities.Artifact {
	t.Helper()
	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    fileName,
		FileType:    contentType,
		FileSize:    1024,
		StorageKey:  "client-1/" + string(artType) + "/test-id.ext",
		Type:        artType,
		Status:      entities.ArtifactStatusPending,
		ContentType: contentType,
	}
	created, err := repo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("failed to create test artifact: %v", err)
	}
	return created
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

// --- Classification Tests ---

func TestConfirmUpload_WithHint_ClassifiesAsHint(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifact(t, artifactRepo)
	fileStorage.PutObject(artifact.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID:      artifact.ID,
		CoachID:         "coach-1",
		DocumentSubtype: entities.DocumentSubtypeInBodyPDF,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Artifact.DocumentSubtype != entities.DocumentSubtypeInBodyPDF {
		t.Errorf("expected document_subtype %s, got %s", entities.DocumentSubtypeInBodyPDF, out.Artifact.DocumentSubtype)
	}
}

func TestConfirmUpload_NoHint_PDFDefaultsToOther(t *testing.T) {
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
		t.Errorf("expected document_subtype %s, got %s", entities.DocumentSubtypeOther, out.Artifact.DocumentSubtype)
	}
}

func TestConfirmUpload_NoHint_AudioClassifiedAsAudio(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifactWithFile(t, artifactRepo, "recording.mp3", "audio/mpeg", entities.ArtifactTypeAudio)
	fileStorage.PutObject(artifact.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Artifact.DocumentSubtype != entities.DocumentSubtypeAudio {
		t.Errorf("expected document_subtype %s, got %s", entities.DocumentSubtypeAudio, out.Artifact.DocumentSubtype)
	}
}

func TestConfirmUpload_MismatchedHintAndExtension_HintWins(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCase()
	// Create a CSV artifact but provide lab_pdf hint
	artifact := createPendingArtifactWithFile(t, artifactRepo, "data.csv", "text/csv", entities.ArtifactTypeDocument)
	fileStorage.PutObject(artifact.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID:      artifact.ID,
		CoachID:         "coach-1",
		DocumentSubtype: entities.DocumentSubtypeLabPDF,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Artifact.DocumentSubtype != entities.DocumentSubtypeLabPDF {
		t.Errorf("expected document_subtype %s, got %s", entities.DocumentSubtypeLabPDF, out.Artifact.DocumentSubtype)
	}
}

// --- Job Enqueue Tests ---

func TestConfirmUpload_EnqueuesExtractionJob(t *testing.T) {
	uc, artifactRepo, fileStorage, _, jobQueue := newConfirmUploadUseCase()
	artifact := createPendingArtifact(t, artifactRepo)
	fileStorage.PutObject(artifact.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID:      artifact.ID,
		CoachID:         "coach-1",
		DocumentSubtype: entities.DocumentSubtypeInBodyPDF,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.JobID == "" {
		t.Fatal("expected non-empty job ID")
	}

	jobs := jobQueue.GetJobs()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Type != entities.JobTypeExtractInBody {
		t.Errorf("expected job type %s, got %s", entities.JobTypeExtractInBody, jobs[0].Type)
	}
}

func TestConfirmUpload_JobPayloadContainsArtifactInfo(t *testing.T) {
	uc, artifactRepo, fileStorage, _, jobQueue := newConfirmUploadUseCase()
	artifact := createPendingArtifact(t, artifactRepo)
	fileStorage.PutObject(artifact.StorageKey)

	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID:      artifact.ID,
		CoachID:         "coach-1",
		DocumentSubtype: entities.DocumentSubtypeLabCSV,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	jobs := jobQueue.GetJobs()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	var payload map[string]string
	if err := json.Unmarshal(jobs[0].Payload, &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload["artifact_id"] != artifact.ID {
		t.Errorf("expected artifact_id %s, got %s", artifact.ID, payload["artifact_id"])
	}
	if payload["document_subtype"] != "lab_csv" {
		t.Errorf("expected document_subtype 'lab_csv', got %s", payload["document_subtype"])
	}
	if payload["storage_key"] != artifact.StorageKey {
		t.Errorf("expected storage_key %s, got %s", artifact.StorageKey, payload["storage_key"])
	}
}

func TestConfirmUpload_OtherSubtype_EnqueuesClassifyJob(t *testing.T) {
	uc, artifactRepo, fileStorage, _, jobQueue := newConfirmUploadUseCase()
	artifact := createPendingArtifact(t, artifactRepo) // PDF without hint → "other"
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
		t.Errorf("expected job type %s, got %s", entities.JobTypeClassifyDocument, jobs[0].Type)
	}
}

func TestConfirmUpload_AudioFile_EnqueuesTranscribeJob(t *testing.T) {
	uc, artifactRepo, fileStorage, _, jobQueue := newConfirmUploadUseCase()
	artifact := createPendingArtifactWithFile(t, artifactRepo, "session.mp3", "audio/mpeg", entities.ArtifactTypeAudio)
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
	if jobs[0].Type != entities.JobTypeTranscribeAudio {
		t.Errorf("expected job type %s, got %s", entities.JobTypeTranscribeAudio, jobs[0].Type)
	}
}

func TestConfirmUpload_ClassificationAuditEventLogged(t *testing.T) {
	uc, artifactRepo, fileStorage, auditRepo, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifact(t, artifactRepo)
	fileStorage.PutObject(artifact.StorageKey)

	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID:      artifact.ID,
		CoachID:         "coach-1",
		DocumentSubtype: entities.DocumentSubtypeInBodyPDF,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	found := false
	for _, e := range events {
		if e.Action == "artifact.classified" {
			found = true
			if e.Metadata["document_subtype"] != string(entities.DocumentSubtypeInBodyPDF) {
				t.Errorf("expected subtype %s in metadata, got %v", entities.DocumentSubtypeInBodyPDF, e.Metadata["document_subtype"])
			}
			break
		}
	}
	if !found {
		t.Error("expected artifact.classified audit event")
	}
}

func TestConfirmUpload_ReturnsJobID(t *testing.T) {
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
	if out.JobID == "" {
		t.Error("expected non-empty job_id in output")
	}
}

func TestConfirmUpload_NoJobQueue_StillSucceeds(t *testing.T) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	auditRepo := repository.NewInMemoryAuditRepository()

	// Create use case without job queue
	uc := NewConfirmUploadUseCase(artifactRepo, fileStorage, auditRepo)

	artifact := createPendingArtifact(t, artifactRepo)
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
	if out.JobID != "" {
		t.Errorf("expected empty job_id when no queue, got %s", out.JobID)
	}
}
