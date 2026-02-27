package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestRequestUpload_WithDocumentSubtype_StoresHint(t *testing.T) {
	uc, artifactRepo, _, _ := newRequestUploadUseCase()

	out, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:        "inbody_scan.pdf",
		FileSize:        1024,
		ContentType:     "application/pdf",
		ClientID:        "client-1",
		CoachID:         "coach-1",
		DocumentSubtype: entities.DocumentSubtypeInBodyPDF,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	artifact, err := artifactRepo.FindByID(context.Background(), out.ArtifactID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if artifact.DocumentSubtype != entities.DocumentSubtypeInBodyPDF {
		t.Errorf("expected document_subtype %s, got %s", entities.DocumentSubtypeInBodyPDF, artifact.DocumentSubtype)
	}
}

func TestRequestUpload_WithoutDocumentSubtype_StoresEmpty(t *testing.T) {
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
	if artifact.DocumentSubtype != "" {
		t.Errorf("expected empty document_subtype, got %s", artifact.DocumentSubtype)
	}
}

func TestRequestUpload_WithInvalidSubtype_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRequestUploadUseCase()

	_, err := uc.Execute(context.Background(), RequestUploadInput{
		FileName:        "report.pdf",
		FileSize:        1024,
		ContentType:     "application/pdf",
		ClientID:        "client-1",
		CoachID:         "coach-1",
		DocumentSubtype: entities.DocumentSubtype("invalid_type"),
	})
	if err == nil {
		t.Fatal("expected error for invalid document subtype")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}
