package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryArtifactRepository_Create_GeneratesID(t *testing.T) {
	repo := NewInMemoryArtifactRepository()
	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "report.pdf",
		FileType:    ".pdf",
		FileSize:    1024,
		StorageKey:  "client-1/document/test.pdf",
		Type:        entities.ArtifactTypeDocument,
		Status:      entities.ArtifactStatusPending,
		ContentType: "application/pdf",
	}

	result, err := repo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID == "" {
		t.Fatal("expected non-empty ID")
	}
}

func TestInMemoryArtifactRepository_Create_SetsTimestamps(t *testing.T) {
	repo := NewInMemoryArtifactRepository()
	art := &entities.Artifact{
		ClientID: "client-1",
		CoachID:  "coach-1",
		FileName: "report.pdf",
		Status:   entities.ArtifactStatusPending,
	}

	result, err := repo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if result.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestInMemoryArtifactRepository_FindByID_ExistingArtifact_ReturnsArtifact(t *testing.T) {
	repo := NewInMemoryArtifactRepository()
	art := &entities.Artifact{
		ClientID: "client-1",
		CoachID:  "coach-1",
		FileName: "report.pdf",
		Status:   entities.ArtifactStatusPending,
	}

	created, err := repo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, err := repo.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil {
		t.Fatal("expected to find artifact")
	}
	if found.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, found.ID)
	}
}

func TestInMemoryArtifactRepository_FindByID_NotFound_ReturnsNil(t *testing.T) {
	repo := NewInMemoryArtifactRepository()

	found, err := repo.FindByID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Fatal("expected nil for nonexistent artifact")
	}
}

func TestInMemoryArtifactRepository_UpdateStatus_Success(t *testing.T) {
	repo := NewInMemoryArtifactRepository()
	art := &entities.Artifact{
		ClientID: "client-1",
		CoachID:  "coach-1",
		FileName: "report.pdf",
		Status:   entities.ArtifactStatusPending,
	}

	created, err := repo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, err := repo.UpdateStatus(context.Background(), created.ID, entities.ArtifactStatusUploaded)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != entities.ArtifactStatusUploaded {
		t.Errorf("expected status %s, got %s", entities.ArtifactStatusUploaded, updated.Status)
	}
}

func TestInMemoryArtifactRepository_UpdateStatus_NotFound_ReturnsError(t *testing.T) {
	repo := NewInMemoryArtifactRepository()

	_, err := repo.UpdateStatus(context.Background(), "nonexistent", entities.ArtifactStatusUploaded)
	if err == nil {
		t.Fatal("expected error for nonexistent artifact")
	}
}

func TestInMemoryArtifactRepository_UpdateDocumentSubtype_Success(t *testing.T) {
	repo := NewInMemoryArtifactRepository()
	art := &entities.Artifact{
		ClientID: "client-1",
		CoachID:  "coach-1",
		FileName: "report.pdf",
		Status:   entities.ArtifactStatusPending,
	}

	created, err := repo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, err := repo.UpdateDocumentSubtype(context.Background(), created.ID, entities.DocumentSubtypeInBodyPDF)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.DocumentSubtype != entities.DocumentSubtypeInBodyPDF {
		t.Errorf("expected subtype %q, got %q", entities.DocumentSubtypeInBodyPDF, updated.DocumentSubtype)
	}
}

func TestInMemoryArtifactRepository_UpdateDocumentSubtype_NotFound_ReturnsError(t *testing.T) {
	repo := NewInMemoryArtifactRepository()

	_, err := repo.UpdateDocumentSubtype(context.Background(), "nonexistent", entities.DocumentSubtypeOther)
	if err == nil {
		t.Fatal("expected error for nonexistent artifact")
	}
}
