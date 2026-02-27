package usecase

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newConfirmUploadUseCaseWithJobQueue() (
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

func createPendingArtifactWithHint(t *testing.T, repo *repository.InMemoryArtifactRepository, hint entities.DocumentSubtype) *entities.Artifact {
	t.Helper()
	art := &entities.Artifact{
		ClientID:        "client-1",
		CoachID:         "coach-1",
		FileName:        "inbody_scan_jan2026.pdf",
		FileType:        "application/pdf",
		FileSize:        1024,
		StorageKey:      "client-1/document/test-id.pdf",
		Type:            entities.ArtifactTypeDocument,
		Status:          entities.ArtifactStatusPending,
		ContentType:     "application/pdf",
		DocumentSubtype: hint,
	}
	created, err := repo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("failed to create test artifact: %v", err)
	}
	return created
}

func TestConfirmUpload_WithHint_ClassifiesAsHint(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCaseWithJobQueue()
	artifact := createPendingArtifactWithHint(t, artifactRepo, entities.DocumentSubtypeInBodyPDF)
	fileStorage.PutObject(artifact.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Artifact.DocumentSubtype != entities.DocumentSubtypeInBodyPDF {
		t.Errorf("expected document_subtype %s, got %s", entities.DocumentSubtypeInBodyPDF, out.Artifact.DocumentSubtype)
	}
}

func TestConfirmUpload_WithoutHint_ClassifiesByExtension(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCaseWithJobQueue()

	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "recording.mp3",
		FileType:    "audio/mpeg",
		FileSize:    5000,
		StorageKey:  "client-1/audio/test-id.mp3",
		Type:        entities.ArtifactTypeAudio,
		Status:      entities.ArtifactStatusPending,
		ContentType: "audio/mpeg",
	}
	created, _ := artifactRepo.Create(context.Background(), art)
	fileStorage.PutObject(created.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: created.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Artifact.DocumentSubtype != entities.DocumentSubtypeAudio {
		t.Errorf("expected document_subtype %s, got %s", entities.DocumentSubtypeAudio, out.Artifact.DocumentSubtype)
	}
}

func TestConfirmUpload_UnknownFileType_ClassifiesAsOther(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCaseWithJobQueue()

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
	created, _ := artifactRepo.Create(context.Background(), art)
	fileStorage.PutObject(created.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: created.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Artifact.DocumentSubtype != entities.DocumentSubtypeOther {
		t.Errorf("expected document_subtype %s, got %s", entities.DocumentSubtypeOther, out.Artifact.DocumentSubtype)
	}
}

func TestConfirmUpload_EnqueuesExtractionJob(t *testing.T) {
	uc, artifactRepo, fileStorage, _, jobQueue := newConfirmUploadUseCaseWithJobQueue()
	artifact := createPendingArtifactWithHint(t, artifactRepo, entities.DocumentSubtypeInBodyPDF)
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
	if len(jobs) == 0 {
		t.Fatal("expected job to be enqueued")
	}
	if jobs[0].Type != entities.JobTypeExtractInBody {
		t.Errorf("expected job type %s, got %s", entities.JobTypeExtractInBody, jobs[0].Type)
	}
}

func TestConfirmUpload_JobPayloadContainsArtifactID(t *testing.T) {
	uc, artifactRepo, fileStorage, _, jobQueue := newConfirmUploadUseCaseWithJobQueue()
	artifact := createPendingArtifactWithHint(t, artifactRepo, entities.DocumentSubtypeLabCSV)
	fileStorage.PutObject(artifact.StorageKey)

	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	jobs := jobQueue.GetJobs()
	if len(jobs) == 0 {
		t.Fatal("expected job to be enqueued")
	}

	var payload map[string]string
	if err := json.Unmarshal(jobs[0].Payload, &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload["artifact_id"] != artifact.ID {
		t.Errorf("expected artifact_id %s in payload, got %s", artifact.ID, payload["artifact_id"])
	}
}

func TestConfirmUpload_OtherType_EnqueuesClassifyJob(t *testing.T) {
	uc, artifactRepo, fileStorage, _, jobQueue := newConfirmUploadUseCaseWithJobQueue()

	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "data.csv",
		FileType:    "text/csv",
		FileSize:    500,
		StorageKey:  "client-1/document/test-id.csv",
		Type:        entities.ArtifactTypeDocument,
		Status:      entities.ArtifactStatusPending,
		ContentType: "text/csv",
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
	if len(jobs) == 0 {
		t.Fatal("expected job to be enqueued")
	}
	if jobs[0].Type != entities.JobTypeClassifyDocument {
		t.Errorf("expected job type %s, got %s", entities.JobTypeClassifyDocument, jobs[0].Type)
	}
}

func TestConfirmUpload_LogsClassificationAuditEvent(t *testing.T) {
	uc, artifactRepo, fileStorage, auditRepo, _ := newConfirmUploadUseCaseWithJobQueue()
	artifact := createPendingArtifactWithHint(t, artifactRepo, entities.DocumentSubtypeInBodyPDF)
	fileStorage.PutObject(artifact.StorageKey)

	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	hasClassified := false
	hasJobEnqueued := false
	for _, e := range events {
		if e.Action == "artifact.classified" {
			hasClassified = true
		}
		if e.Action == "artifact.job_enqueued" {
			hasJobEnqueued = true
		}
	}
	if !hasClassified {
		t.Error("expected artifact.classified audit event")
	}
	if !hasJobEnqueued {
		t.Error("expected artifact.job_enqueued audit event")
	}
}

func TestConfirmUpload_MismatchedHintAndExtension_HintWins(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newConfirmUploadUseCaseWithJobQueue()

	art := &entities.Artifact{
		ClientID:        "client-1",
		CoachID:         "coach-1",
		FileName:        "data.csv",
		FileType:        "text/csv",
		FileSize:        500,
		StorageKey:      "client-1/document/test-id.csv",
		Type:            entities.ArtifactTypeDocument,
		Status:          entities.ArtifactStatusPending,
		ContentType:     "text/csv",
		DocumentSubtype: entities.DocumentSubtypeWearableCSV,
	}
	created, _ := artifactRepo.Create(context.Background(), art)
	fileStorage.PutObject(created.StorageKey)

	out, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: created.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Artifact.DocumentSubtype != entities.DocumentSubtypeWearableCSV {
		t.Errorf("expected hint %s to win, got %s", entities.DocumentSubtypeWearableCSV, out.Artifact.DocumentSubtype)
	}
}
