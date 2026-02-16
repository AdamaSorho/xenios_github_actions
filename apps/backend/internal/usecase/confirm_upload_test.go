package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newConfirmUploadUseCase() (*ConfirmUploadUseCase, *repository.InMemoryArtifactRepository, *repository.InMemoryFileStorage, *repository.InMemoryAuditRepository) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := NewConfirmUploadUseCase(artifactRepo, fileStorage, auditRepo)
	return uc, artifactRepo, fileStorage, auditRepo
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

func TestConfirmUpload_ValidInput_UpdatesStatusToUploaded(t *testing.T) {
	uc, artifactRepo, fileStorage, _ := newConfirmUploadUseCase()
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
	uc, _, _, _ := newConfirmUploadUseCase()

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
	uc, _, _, _ := newConfirmUploadUseCase()

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
	uc, _, _, _ := newConfirmUploadUseCase()

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
	uc, artifactRepo, fileStorage, _ := newConfirmUploadUseCase()
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
	uc, artifactRepo, fileStorage, _ := newConfirmUploadUseCase()
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
	uc, artifactRepo, _, _ := newConfirmUploadUseCase()
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
	uc, artifactRepo, _, _ := newConfirmUploadUseCase()
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
	uc, artifactRepo, fileStorage, auditRepo := newConfirmUploadUseCase()
	artifact := createPendingArtifact(t, artifactRepo)
	fileStorage.PutObject(artifact.StorageKey)

	_, err := uc.Execute(context.Background(), ConfirmUploadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(auditRepo.Events) == 0 {
		t.Fatal("expected audit event to be logged")
	}
	if auditRepo.Events[0].Action != "artifact.upload_confirmed" {
		t.Errorf("expected action 'artifact.upload_confirmed', got '%s'", auditRepo.Events[0].Action)
	}
}
