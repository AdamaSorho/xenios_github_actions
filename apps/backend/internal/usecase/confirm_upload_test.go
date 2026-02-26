package usecase

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newConfirmUploadUseCase() (
	*ConfirmUploadUseCase,
	*repository.InMemoryArtifactRepository,
	*repository.InMemoryFileStorage,
	*repository.InMemoryAuditRepository,
	*repository.InMemoryJobQueue,
) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	auditRepo := repository.NewInMemoryAuditRepository()
	jobQueue := repository.NewInMemoryJobQueue()

	uc := NewConfirmUploadUseCase(artifactRepo, fileStorage, auditRepo, jobQueue)
	return uc, artifactRepo, fileStorage, auditRepo, jobQueue
}

func createPendingArtifact(t *testing.T, repo *repository.InMemoryArtifactRepository) *entities.Artifact {
	t.Helper()
	return createPendingArtifactWithSubtype(t, repo, "report.pdf", "application/pdf", "")
}

func createPendingArtifactWithSubtype(t *testing.T, repo *repository.InMemoryArtifactRepository, fileName, contentType string, subtype entities.DocumentSubtype) *entities.Artifact {
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
		DocumentSubtype: subtype,
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

	_, _ = uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})

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
	if len(events) < 1 {
		t.Fatal("expected at least one audit event")
	}
	if events[0].Action != "artifact.upload_confirmed" {
		t.Errorf("expected action 'artifact.upload_confirmed', got '%s'", events[0].Action)
	}
}

// --- Classification tests ---

func TestConfirmUpload_WithHint_ClassifiesUsingHint(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifactWithSubtype(t, artifactRepo, "scan.pdf", "application/pdf", entities.DocumentSubtypeInBodyPDF)
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

func TestConfirmUpload_NoHint_ClassifiesByExtension(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCase()

	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "recording.mp3",
		FileType:    "audio/mpeg",
		FileSize:    1024,
		StorageKey:  "client-1/audio/test-id.mp3",
		Type:        entities.ArtifactTypeAudio,
		Status:      entities.ArtifactStatusPending,
		ContentType: "audio/mpeg",
	}
	created, err := artifactRepo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("create artifact: %v", err)
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

func TestConfirmUpload_NoHint_UnknownType_ClassifiesAsOther(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifact(t, artifactRepo) // report.pdf with no hint
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

// --- Job enqueueing tests ---

func TestConfirmUpload_EnqueuesExtractionJob(t *testing.T) {
	uc, artifactRepo, fileStorage, _, jobQueue := newConfirmUploadUseCase()
	artifact := createPendingArtifactWithSubtype(t, artifactRepo, "scan.pdf", "application/pdf", entities.DocumentSubtypeInBodyPDF)
	fileStorage.PutObject(artifact.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.JobID == "" {
		t.Error("expected non-empty job ID")
	}

	jobs := jobQueue.GetJobs()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job enqueued, got %d", len(jobs))
	}
	if jobs[0].Type != entities.JobTypeExtractInBody {
		t.Errorf("expected job type %q, got %q", entities.JobTypeExtractInBody, jobs[0].Type)
	}
}

func TestConfirmUpload_JobPayloadContainsArtifactInfo(t *testing.T) {
	uc, artifactRepo, fileStorage, _, jobQueue := newConfirmUploadUseCase()
	artifact := createPendingArtifactWithSubtype(t, artifactRepo, "scan.pdf", "application/pdf", entities.DocumentSubtypeInBodyPDF)
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

	var payload ExtractionJobPayload
	if err := json.Unmarshal(jobs[0].Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.ArtifactID != artifact.ID {
		t.Errorf("expected artifact_id %q, got %q", artifact.ID, payload.ArtifactID)
	}
	if payload.DocumentSubtype != string(entities.DocumentSubtypeInBodyPDF) {
		t.Errorf("expected document_subtype %q, got %q", entities.DocumentSubtypeInBodyPDF, payload.DocumentSubtype)
	}
}

func TestConfirmUpload_OtherType_EnqueuesClassifyDocumentJob(t *testing.T) {
	uc, artifactRepo, fileStorage, _, jobQueue := newConfirmUploadUseCase()
	artifact := createPendingArtifact(t, artifactRepo) // PDF with no hint -> "other"
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

func TestConfirmUpload_AudioFile_EnqueuesTranscribeAudioJob(t *testing.T) {
	uc, artifactRepo, fileStorage, _, jobQueue := newConfirmUploadUseCase()

	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "session.mp3",
		FileType:    "audio/mpeg",
		FileSize:    5000,
		StorageKey:  "client-1/audio/test.mp3",
		Type:        entities.ArtifactTypeAudio,
		Status:      entities.ArtifactStatusPending,
		ContentType: "audio/mpeg",
	}
	created, _ := artifactRepo.Create(context.Background(), art)
	fileStorage.PutObject(created.StorageKey)

	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
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
	artifact := createPendingArtifactWithSubtype(t, artifactRepo, "scan.pdf", "application/pdf", entities.DocumentSubtypeInBodyPDF)
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

	// Find the classification event
	found := false
	for _, e := range events {
		if e.Action == "artifact.classified" {
			found = true
			if e.Metadata["document_subtype"] != string(entities.DocumentSubtypeInBodyPDF) {
				t.Errorf("expected document_subtype metadata %q, got %q", entities.DocumentSubtypeInBodyPDF, e.Metadata["document_subtype"])
			}
		}
	}
	if !found {
		t.Error("expected 'artifact.classified' audit event")
	}
}
