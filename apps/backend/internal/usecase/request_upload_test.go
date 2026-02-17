package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
)

func newRequestUploadUseCase() (*RequestUploadUseCase, *repository.InMemoryArtifactRepository, *repository.InMemoryFileStorage, *repository.InMemoryAuditRepository) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := NewRequestUploadUseCase(artifactRepo, fileStorage, auditRepo)
	return uc, artifactRepo, fileStorage, auditRepo
}

func TestRequestUpload_ValidInput_ReturnsPresignedURL(t *testing.T) {
	uc, _, _, _ := newRequestUploadUseCase()

	out, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:    "report.pdf",
		FileSize:    1024,
		ContentType: "application/pdf",
		ClientID:    "client-1",
		CoachID:     "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.PresignedURL == "" {
		t.Error("expected non-empty presigned URL")
	}
	if out.ArtifactID == "" {
		t.Error("expected non-empty artifact ID")
	}
	if out.ExpiresAt.IsZero() {
		t.Error("expected non-zero expiry")
	}
}

func TestRequestUpload_ValidInput_CreatesArtifactWithPendingStatus(t *testing.T) {
	uc, artifactRepo, _, _ := newRequestUploadUseCase()

	out, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:    "report.pdf",
		FileSize:    1024,
		ContentType: "application/pdf",
		ClientID:    "client-1",
		CoachID:     "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	artifact, err := artifactRepo.FindByID(context.Background(), out.ArtifactID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if artifact == nil {
		t.Fatal("expected artifact to be persisted")
	}
	if artifact.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", artifact.Status)
	}
}

func TestRequestUpload_ValidInput_SetsCorrectStorageKey(t *testing.T) {
	uc, _, _, _ := newRequestUploadUseCase()

	out, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:    "recording.mp3",
		FileSize:    5 * 1024 * 1024,
		ContentType: "audio/mpeg",
		ClientID:    "client-123",
		CoachID:     "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.StorageKey == "" {
		t.Error("expected non-empty storage key")
	}
	// Key should start with client ID and contain artifact type
	expected := "client-123/audio/"
	if len(out.StorageKey) < len(expected) || out.StorageKey[:len(expected)] != expected {
		t.Errorf("expected storage key to start with '%s', got '%s'", expected, out.StorageKey)
	}
}

func TestRequestUpload_EmptyFileName_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRequestUploadUseCase()

	_, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:    "",
		FileSize:    1024,
		ContentType: "application/pdf",
		ClientID:    "client-1",
		CoachID:     "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestRequestUpload_EmptyContentType_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRequestUploadUseCase()

	_, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:    "report.pdf",
		FileSize:    1024,
		ContentType: "",
		ClientID:    "client-1",
		CoachID:     "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestRequestUpload_EmptyClientID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRequestUploadUseCase()

	_, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:    "report.pdf",
		FileSize:    1024,
		ContentType: "application/pdf",
		ClientID:    "",
		CoachID:     "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestRequestUpload_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRequestUploadUseCase()

	_, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:    "report.pdf",
		FileSize:    1024,
		ContentType: "application/pdf",
		ClientID:    "client-1",
		CoachID:     "",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestRequestUpload_DisallowedFileType_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRequestUploadUseCase()

	_, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:    "script.exe",
		FileSize:    1024,
		ContentType: "application/octet-stream",
		ClientID:    "client-1",
		CoachID:     "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestRequestUpload_FileTooLarge_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRequestUploadUseCase()

	_, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:    "report.pdf",
		FileSize:    11 * 1024 * 1024, // 11MB exceeds 10MB limit for documents
		ContentType: "application/pdf",
		ClientID:    "client-1",
		CoachID:     "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestRequestUpload_AudioFileLargeButWithinLimit_Succeeds(t *testing.T) {
	uc, _, _, _ := newRequestUploadUseCase()

	out, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:    "recording.mp3",
		FileSize:    50 * 1024 * 1024, // 50MB within 100MB limit
		ContentType: "audio/mpeg",
		ClientID:    "client-1",
		CoachID:     "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.PresignedURL == "" {
		t.Error("expected non-empty presigned URL")
	}
}

func TestRequestUpload_InvalidContentType_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRequestUploadUseCase()

	_, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:    "report.pdf",
		FileSize:    1024,
		ContentType: "text/html",
		ClientID:    "client-1",
		CoachID:     "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestRequestUpload_AuditEventLogged(t *testing.T) {
	uc, _, _, auditRepo := newRequestUploadUseCase()

	_, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:    "report.pdf",
		FileSize:    1024,
		ContentType: "application/pdf",
		ClientID:    "client-1",
		CoachID:     "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	if len(events) == 0 {
		t.Fatal("expected audit event to be logged")
	}
	if events[0].Action != "artifact.upload_requested" {
		t.Errorf("expected action 'artifact.upload_requested', got '%s'", events[0].Action)
	}
}

func TestRequestUpload_ZeroFileSize_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRequestUploadUseCase()

	_, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:    "report.pdf",
		FileSize:    0,
		ContentType: "application/pdf",
		ClientID:    "client-1",
		CoachID:     "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

// --- DocumentSubtype hint tests ---

func TestRequestUpload_WithValidDocumentSubtypeHint_StoresHint(t *testing.T) {
	uc, artifactRepo, _, _ := newRequestUploadUseCase()

	out, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:        "inbody_scan.pdf",
		FileSize:        1024,
		ContentType:     "application/pdf",
		ClientID:        "client-1",
		CoachID:         "coach-1",
		DocumentSubtype: "inbody_pdf",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	artifact, err := artifactRepo.FindByID(context.Background(), out.ArtifactID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if artifact.DocumentSubtype != "inbody_pdf" {
		t.Errorf("expected document_subtype 'inbody_pdf', got '%s'", artifact.DocumentSubtype)
	}
}

func TestRequestUpload_WithEmptyDocumentSubtype_Succeeds(t *testing.T) {
	uc, _, _, _ := newRequestUploadUseCase()

	out, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:        "report.pdf",
		FileSize:        1024,
		ContentType:     "application/pdf",
		ClientID:        "client-1",
		CoachID:         "coach-1",
		DocumentSubtype: "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ArtifactID == "" {
		t.Error("expected non-empty artifact ID")
	}
}

func TestRequestUpload_WithInvalidDocumentSubtype_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRequestUploadUseCase()

	_, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:        "report.pdf",
		FileSize:        1024,
		ContentType:     "application/pdf",
		ClientID:        "client-1",
		CoachID:         "coach-1",
		DocumentSubtype: "invalid_subtype",
	})
	if err == nil {
		t.Fatal("expected error for invalid document_subtype")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}
