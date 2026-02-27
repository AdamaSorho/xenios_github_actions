package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/xenios/backend/internal/adapter/middleware"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

var errInternal = errors.New("internal error")

func TestUploadHandler_RequestPresignedURL_WithDocumentSubtype_Success(t *testing.T) {
	h := NewUploadHandler(
		&mockRequestUploadUC{
			output: &usecase.RequestUploadOutput{
				PresignedURL: "https://s3.example.com/presigned",
				ArtifactID:   "art-1",
				ExpiresAt:    time.Now().Add(15 * time.Minute),
				StorageKey:   "client-1/document/art-1.pdf",
				Artifact: &entities.Artifact{
					ID:              "art-1",
					ClientID:        "client-1",
					CoachID:         "coach-1",
					FileName:        "inbody_scan.pdf",
					Status:          entities.ArtifactStatusPending,
					DocumentSubtype: entities.DocumentSubtypeInBodyPDF,
				},
			},
		},
		&mockConfirmUploadUC{},
		&mockRequestDownloadUC{},
	)

	body, _ := json.Marshal(PresignRequest{
		FileName:        "inbody_scan.pdf",
		FileSize:        1024,
		ContentType:     "application/pdf",
		ClientID:        "client-1",
		DocumentSubtype: string(entities.DocumentSubtypeInBodyPDF),
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/presign", bytes.NewReader(body))
	req = withAuth(req)

	h.RequestPresignedURL(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp usecase.RequestUploadOutput
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Artifact.DocumentSubtype != entities.DocumentSubtypeInBodyPDF {
		t.Errorf("expected document_subtype %s, got %s", entities.DocumentSubtypeInBodyPDF, resp.Artifact.DocumentSubtype)
	}
}

func TestUploadHandler_ConfirmUpload_ReturnsJobID(t *testing.T) {
	h := NewUploadHandler(
		&mockRequestUploadUC{},
		&mockConfirmUploadUC{
			output: &usecase.ConfirmUploadOutput{
				Artifact: &entities.Artifact{
					ID:              "art-1",
					ClientID:        "client-1",
					CoachID:         "coach-1",
					FileName:        "inbody_scan.pdf",
					Status:          entities.ArtifactStatusUploaded,
					DocumentSubtype: entities.DocumentSubtypeInBodyPDF,
				},
				JobID: "job-123",
			},
		},
		&mockRequestDownloadUC{},
	)

	r := chi.NewRouter()
	r.Post("/api/v1/uploads/{artifactID}/confirm", h.ConfirmUpload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/art-1/confirm", nil)
	ctx := middleware.SetUserClaims(req.Context(), &middleware.UserClaims{
		Subject: "coach-1",
		Role:    "coach",
	})
	req = req.WithContext(ctx)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp usecase.ConfirmUploadOutput
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.JobID != "job-123" {
		t.Errorf("expected job_id 'job-123', got '%s'", resp.JobID)
	}
}

func TestUploadHandler_ConfirmUpload_InternalError_Returns500(t *testing.T) {
	h := NewUploadHandler(
		&mockRequestUploadUC{},
		&mockConfirmUploadUC{err: errInternal},
		&mockRequestDownloadUC{},
	)

	r := chi.NewRouter()
	r.Post("/api/v1/uploads/{artifactID}/confirm", h.ConfirmUpload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/art-1/confirm", nil)
	ctx := middleware.SetUserClaims(req.Context(), &middleware.UserClaims{
		Subject: "coach-1",
		Role:    "coach",
	})
	req = req.WithContext(ctx)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}
