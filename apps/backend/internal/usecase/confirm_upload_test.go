package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newConfirmUploadUseCase() (*ConfirmUploadUseCase, *repository.InMemoryArtifactRepository, *repository.InMemoryFileStorage, *repository.InMemoryAuditRepository, *mockJobQueue) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	auditRepo := repository.NewInMemoryAuditRepository()
	jq := &mockJobQueue{
		enqueueFunc: func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
			return &entities.Job{
				ID:        "job-123",
				Type:      jobType,
				Status:    entities.JobStatusCreated,
				CreatedAt: time.Now(),
			}, nil
		},
	}

	uc := NewConfirmUploadUseCase(artifactRepo, fileStorage, auditRepo, jq)
	return uc, artifactRepo, fileStorage, auditRepo, jq
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

func createPendingArtifactWithHint(t *testing.T, repo *repository.InMemoryArtifactRepository, fileName, contentType string, hint entities.DocumentSubtype) *entities.Artifact {
	t.Helper()
	artType, _ := entities.ValidateFileExtension(fileName)
	art := &entities.Artifact{
		ClientID:        "client-1",
		CoachID:         "coach-1",
		FileName:        fileName,
		FileType:        contentType,
		FileSize:        1024,
		StorageKey:      "client-1/" + string(artType) + "/test-id.pdf",
		Type:            artType,
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

// --- New tests for classification and job enqueuing ---

func TestConfirmUpload_ClassifiesFileType_PDFWithHint(t *testing.T) {
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

func TestConfirmUpload_ClassifiesFileType_AudioNoHint(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifactWithHint(t, artifactRepo, "session.mp3", "audio/mpeg", "")
	fileStorage.PutObject(artifact.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Artifact.DocumentSubtype != entities.DocumentSubtypeAudio {
		t.Errorf("expected document_subtype %q, got %q", entities.DocumentSubtypeAudio, out.Artifact.DocumentSubtype)
	}
}

func TestConfirmUpload_ClassifiesFileType_CSVNoHint_ReturnsOther(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifactWithHint(t, artifactRepo, "data.csv", "text/csv", "")
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

func TestConfirmUpload_ClassifiesFileType_JSONNoHint_ReturnsWearableJSON(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifactWithHint(t, artifactRepo, "export.json", "application/json", "")
	fileStorage.PutObject(artifact.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Artifact.DocumentSubtype != entities.DocumentSubtypeWearableJSON {
		t.Errorf("expected document_subtype %q, got %q", entities.DocumentSubtypeWearableJSON, out.Artifact.DocumentSubtype)
	}
}

func TestConfirmUpload_EnqueuesExtractionJob_InBodyPDF(t *testing.T) {
	var enqueuedJobType entities.JobType
	var enqueuedPayload []byte

	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	auditRepo := repository.NewInMemoryAuditRepository()
	jq := &mockJobQueue{
		enqueueFunc: func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
			enqueuedJobType = jobType
			enqueuedPayload = payload
			return &entities.Job{ID: "job-456", Type: jobType, Status: entities.JobStatusCreated}, nil
		},
	}

	uc := NewConfirmUploadUseCase(artifactRepo, fileStorage, auditRepo, jq)
	artifact := createPendingArtifactWithHint(t, artifactRepo, "scan.pdf", "application/pdf", entities.DocumentSubtypeInBodyPDF)
	fileStorage.PutObject(artifact.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enqueuedJobType != entities.JobTypeExtractInBody {
		t.Errorf("expected job type %q, got %q", entities.JobTypeExtractInBody, enqueuedJobType)
	}
	if out.JobID != "job-456" {
		t.Errorf("expected job ID 'job-456', got %q", out.JobID)
	}
	if len(enqueuedPayload) == 0 {
		t.Error("expected non-empty job payload")
	}
}

func TestConfirmUpload_EnqueuesExtractionJob_AudioFile(t *testing.T) {
	var enqueuedJobType entities.JobType

	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	auditRepo := repository.NewInMemoryAuditRepository()
	jq := &mockJobQueue{
		enqueueFunc: func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
			enqueuedJobType = jobType
			return &entities.Job{ID: "job-789", Type: jobType, Status: entities.JobStatusCreated}, nil
		},
	}

	uc := NewConfirmUploadUseCase(artifactRepo, fileStorage, auditRepo, jq)
	artifact := createPendingArtifactWithHint(t, artifactRepo, "session.mp3", "audio/mpeg", "")
	fileStorage.PutObject(artifact.StorageKey)

	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enqueuedJobType != entities.JobTypeTranscribeAudio {
		t.Errorf("expected job type %q, got %q", entities.JobTypeTranscribeAudio, enqueuedJobType)
	}
}

func TestConfirmUpload_EnqueuesExtractionJob_UnknownType_ClassifyDocument(t *testing.T) {
	var enqueuedJobType entities.JobType

	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	auditRepo := repository.NewInMemoryAuditRepository()
	jq := &mockJobQueue{
		enqueueFunc: func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
			enqueuedJobType = jobType
			return &entities.Job{ID: "job-classify", Type: jobType, Status: entities.JobStatusCreated}, nil
		},
	}

	uc := NewConfirmUploadUseCase(artifactRepo, fileStorage, auditRepo, jq)
	artifact := createPendingArtifactWithHint(t, artifactRepo, "unknown.pdf", "application/pdf", "")
	fileStorage.PutObject(artifact.StorageKey)

	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enqueuedJobType != entities.JobTypeClassifyDocument {
		t.Errorf("expected job type %q, got %q", entities.JobTypeClassifyDocument, enqueuedJobType)
	}
}

func TestConfirmUpload_ClassificationAuditEventLogged(t *testing.T) {
	uc, artifactRepo, fileStorage, auditRepo, _ := newConfirmUploadUseCase()
	artifact := createPendingArtifactWithHint(t, artifactRepo, "scan.pdf", "application/pdf", entities.DocumentSubtypeInBodyPDF)
	fileStorage.PutObject(artifact.StorageKey)

	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	// Should have: upload_confirmed, classified, job_enqueued
	var foundClassified, foundJobEnqueued bool
	for _, e := range events {
		if e.Action == "artifact.classified" {
			foundClassified = true
			if e.Metadata["document_subtype"] != string(entities.DocumentSubtypeInBodyPDF) {
				t.Errorf("expected document_subtype metadata %q, got %v", entities.DocumentSubtypeInBodyPDF, e.Metadata["document_subtype"])
			}
		}
		if e.Action == "artifact.job_enqueued" {
			foundJobEnqueued = true
			if e.Metadata["job_type"] != string(entities.JobTypeExtractInBody) {
				t.Errorf("expected job_type metadata %q, got %v", entities.JobTypeExtractInBody, e.Metadata["job_type"])
			}
		}
	}
	if !foundClassified {
		t.Error("expected 'artifact.classified' audit event to be logged")
	}
	if !foundJobEnqueued {
		t.Error("expected 'artifact.job_enqueued' audit event to be logged")
	}
}

func TestConfirmUpload_NilJobQueue_SkipsEnqueueGracefully(t *testing.T) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	auditRepo := repository.NewInMemoryAuditRepository()

	// Pass nil for job queue
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
		t.Errorf("expected empty job ID when queue is nil, got %q", out.JobID)
	}
	if out.Artifact.Status != entities.ArtifactStatusUploaded {
		t.Errorf("expected status 'uploaded', got '%s'", out.Artifact.Status)
	}
}

func TestConfirmUpload_OutputIncludesJobID(t *testing.T) {
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
		t.Error("expected non-empty job ID")
	}
}
