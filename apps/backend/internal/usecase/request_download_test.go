package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newRequestDownloadUseCase() (*RequestDownloadUseCase, *repository.InMemoryArtifactRepository, *repository.InMemoryFileStorage, *repository.InMemoryAuditRepository) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := NewRequestDownloadUseCase(artifactRepo, fileStorage, auditRepo)
	return uc, artifactRepo, fileStorage, auditRepo
}

func createUploadedArtifact(t *testing.T, repo *repository.InMemoryArtifactRepository, fileStorage *repository.InMemoryFileStorage) *entities.Artifact {
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

	// Simulate upload
	fileStorage.PutObject(created.StorageKey)
	updated, err := repo.UpdateStatus(context.Background(), created.ID, entities.ArtifactStatusUploaded)
	if err != nil {
		t.Fatalf("failed to update artifact status: %v", err)
	}
	return updated
}

func TestRequestDownload_ValidInput_ReturnsPresignedURL(t *testing.T) {
	uc, artifactRepo, fileStorage, _ := newRequestDownloadUseCase()
	artifact := createUploadedArtifact(t, artifactRepo, fileStorage)

	out, err := uc.Execute(context.Background(), RequestDownloadInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.PresignedURL == "" {
		t.Error("expected non-empty presigned URL")
	}
	if out.ExpiresAt.IsZero() {
		t.Error("expected non-zero expiry")
	}
}

func TestRequestDownload_EmptyArtifactID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRequestDownloadUseCase()

	_, err := uc.Execute(context.Background(), RequestDownloadInput{
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

func TestRequestDownload_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRequestDownloadUseCase()

	_, err := uc.Execute(context.Background(), RequestDownloadInput{
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

func TestRequestDownload_ArtifactNotFound_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRequestDownloadUseCase()

	_, err := uc.Execute(context.Background(), RequestDownloadInput{
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

func TestRequestDownload_WrongCoach_ReturnsAuthError(t *testing.T) {
	uc, artifactRepo, fileStorage, _ := newRequestDownloadUseCase()
	artifact := createUploadedArtifact(t, artifactRepo, fileStorage)

	_, err := uc.Execute(context.Background(), RequestDownloadInput{
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

func TestRequestDownload_PendingArtifact_ReturnsValidationError(t *testing.T) {
	uc, artifactRepo, _, _ := newRequestDownloadUseCase()

	// Create a pending artifact (not uploaded)
	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "report.pdf",
		StorageKey:  "client-1/document/test.pdf",
		Status:      entities.ArtifactStatusPending,
		ContentType: "application/pdf",
	}
	created, err := artifactRepo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("failed to create test artifact: %v", err)
	}

	_, err = uc.Execute(context.Background(), RequestDownloadInput{
		ArtifactID: created.ID,
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error for pending artifact")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestRequestDownload_FailedArtifact_ReturnsValidationError(t *testing.T) {
	uc, artifactRepo, _, _ := newRequestDownloadUseCase()

	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "report.pdf",
		StorageKey:  "client-1/document/test.pdf",
		Status:      entities.ArtifactStatusPending,
		ContentType: "application/pdf",
	}
	created, err := artifactRepo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("failed to create test artifact: %v", err)
	}
	_, err = artifactRepo.UpdateStatus(context.Background(), created.ID, entities.ArtifactStatusFailed)
	if err != nil {
		t.Fatalf("failed to update artifact status: %v", err)
	}

	_, err = uc.Execute(context.Background(), RequestDownloadInput{
		ArtifactID: created.ID,
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error for failed artifact")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestRequestDownload_AuditEventLogged(t *testing.T) {
	uc, artifactRepo, fileStorage, auditRepo := newRequestDownloadUseCase()
	artifact := createUploadedArtifact(t, artifactRepo, fileStorage)

	_, err := uc.Execute(context.Background(), RequestDownloadInput{
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
	if events[0].Action != "artifact.download_requested" {
		t.Errorf("expected action 'artifact.download_requested', got '%s'", events[0].Action)
	}
}
