package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/xenios/backend/internal/adapter/middleware"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// --- Mock use cases ---

type mockRequestUploadUC struct {
	output *usecase.RequestUploadOutput
	err    error
}

func (m *mockRequestUploadUC) Execute(_ context.Context, _ usecase.RequestUploadInput) (*usecase.RequestUploadOutput, error) {
	return m.output, m.err
}

type mockConfirmUploadUC struct {
	output   *usecase.ConfirmUploadOutput
	err      error
	lastInput usecase.ConfirmUploadInput
}

func (m *mockConfirmUploadUC) Execute(_ context.Context, input usecase.ConfirmUploadInput) (*usecase.ConfirmUploadOutput, error) {
	m.lastInput = input
	return m.output, m.err
}

type mockRequestDownloadUC struct {
	output *usecase.RequestDownloadOutput
	err    error
}

func (m *mockRequestDownloadUC) Execute(_ context.Context, _ usecase.RequestDownloadInput) (*usecase.RequestDownloadOutput, error) {
	return m.output, m.err
}

func defaultConfirmUploadMock() *mockConfirmUploadUC {
	return &mockConfirmUploadUC{
		output: &usecase.ConfirmUploadOutput{
			Artifact: &entities.Artifact{
				ID:              "art-1",
				ClientID:        "client-1",
				CoachID:         "coach-1",
				FileName:        "report.pdf",
				Status:          entities.ArtifactStatusUploaded,
				DocumentSubtype: entities.DocumentSubtypeOther,
			},
			JobID: "job-1",
		},
	}
}

func defaultUploadHandler() *UploadHandler {
	return NewUploadHandler(
		&mockRequestUploadUC{
			output: &usecase.RequestUploadOutput{
				PresignedURL: "https://s3.example.com/presigned",
				ArtifactID:   "art-1",
				ExpiresAt:    time.Now().Add(15 * time.Minute),
				StorageKey:   "client-1/document/art-1.pdf",
				Artifact: &entities.Artifact{
					ID:       "art-1",
					ClientID: "client-1",
					CoachID:  "coach-1",
					FileName: "report.pdf",
					Status:   entities.ArtifactStatusPending,
				},
			},
		},
		defaultConfirmUploadMock(),
		&mockRequestDownloadUC{
			output: &usecase.RequestDownloadOutput{
				PresignedURL: "https://s3.example.com/download",
				ExpiresAt:    time.Now().Add(15 * time.Minute),
				Artifact: &entities.Artifact{
					ID:       "art-1",
					ClientID: "client-1",
					CoachID:  "coach-1",
					FileName: "report.pdf",
					Status:   entities.ArtifactStatusUploaded,
				},
			},
		},
	)
}

func withAuth(r *http.Request) *http.Request {
	ctx := middleware.SetUserClaims(r.Context(), &middleware.UserClaims{
		Subject: "coach-1",
		Role:    "coach",
	})
	return r.WithContext(ctx)
}

// --- RequestPresignedURL tests ---

func TestUploadHandler_RequestPresignedURL_Success(t *testing.T) {
	h := defaultUploadHandler()

	body, _ := json.Marshal(PresignRequest{
		FileName:    "report.pdf",
		FileSize:    1024,
		ContentType: "application/pdf",
		ClientID:    "client-1",
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
	if resp.PresignedURL == "" {
		t.Error("expected non-empty presigned URL")
	}
	if resp.ArtifactID == "" {
		t.Error("expected non-empty artifact ID")
	}
}

func TestUploadHandler_RequestPresignedURL_NoAuth_Returns401(t *testing.T) {
	h := defaultUploadHandler()

	body, _ := json.Marshal(PresignRequest{
		FileName:    "report.pdf",
		FileSize:    1024,
		ContentType: "application/pdf",
		ClientID:    "client-1",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/presign", bytes.NewReader(body))
	// No auth context

	h.RequestPresignedURL(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestUploadHandler_RequestPresignedURL_InvalidJSON_Returns400(t *testing.T) {
	h := defaultUploadHandler()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/presign", bytes.NewReader([]byte("not json")))
	req = withAuth(req)

	h.RequestPresignedURL(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestUploadHandler_RequestPresignedURL_ValidationError_Returns400(t *testing.T) {
	h := NewUploadHandler(
		&mockRequestUploadUC{err: &usecase.ValidationError{Message: "file_name is required"}},
		&mockConfirmUploadUC{},
		&mockRequestDownloadUC{},
	)

	body, _ := json.Marshal(PresignRequest{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/presign", bytes.NewReader(body))
	req = withAuth(req)

	h.RequestPresignedURL(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// --- ConfirmUpload tests ---

func TestUploadHandler_ConfirmUpload_Success(t *testing.T) {
	h := defaultUploadHandler()

	r := chi.NewRouter()
	r.Post("/api/v1/uploads/{artifactID}/confirm", h.ConfirmUpload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/art-1/confirm", nil)
	req = withAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestUploadHandler_ConfirmUpload_NoAuth_Returns401(t *testing.T) {
	h := defaultUploadHandler()

	r := chi.NewRouter()
	r.Post("/api/v1/uploads/{artifactID}/confirm", h.ConfirmUpload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/art-1/confirm", nil)
	// No auth context

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestUploadHandler_ConfirmUpload_AuthError_Returns403(t *testing.T) {
	h := NewUploadHandler(
		&mockRequestUploadUC{},
		&mockConfirmUploadUC{err: &usecase.AuthenticationError{Message: "not authorized"}},
		&mockRequestDownloadUC{},
	)

	r := chi.NewRouter()
	r.Post("/api/v1/uploads/{artifactID}/confirm", h.ConfirmUpload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/art-1/confirm", nil)
	req = withAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestUploadHandler_ConfirmUpload_ValidationError_Returns400(t *testing.T) {
	h := NewUploadHandler(
		&mockRequestUploadUC{},
		&mockConfirmUploadUC{err: &usecase.ValidationError{Message: "artifact not found"}},
		&mockRequestDownloadUC{},
	)

	r := chi.NewRouter()
	r.Post("/api/v1/uploads/{artifactID}/confirm", h.ConfirmUpload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/art-1/confirm", nil)
	req = withAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// --- RequestDownloadURL tests ---

func TestUploadHandler_RequestDownloadURL_Success(t *testing.T) {
	h := defaultUploadHandler()

	r := chi.NewRouter()
	r.Post("/api/v1/uploads/{artifactID}/download", h.RequestDownloadURL)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/art-1/download", nil)
	req = withAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp usecase.RequestDownloadOutput
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.PresignedURL == "" {
		t.Error("expected non-empty presigned URL")
	}
}

func TestUploadHandler_RequestDownloadURL_NoAuth_Returns401(t *testing.T) {
	h := defaultUploadHandler()

	r := chi.NewRouter()
	r.Post("/api/v1/uploads/{artifactID}/download", h.RequestDownloadURL)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/art-1/download", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestUploadHandler_RequestDownloadURL_AuthError_Returns403(t *testing.T) {
	h := NewUploadHandler(
		&mockRequestUploadUC{},
		&mockConfirmUploadUC{},
		&mockRequestDownloadUC{err: &usecase.AuthenticationError{Message: "not authorized"}},
	)

	r := chi.NewRouter()
	r.Post("/api/v1/uploads/{artifactID}/download", h.RequestDownloadURL)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/art-1/download", nil)
	req = withAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestUploadHandler_RequestDownloadURL_ValidationError_Returns400(t *testing.T) {
	h := NewUploadHandler(
		&mockRequestUploadUC{},
		&mockConfirmUploadUC{},
		&mockRequestDownloadUC{err: &usecase.ValidationError{Message: "artifact not found"}},
	)

	r := chi.NewRouter()
	r.Post("/api/v1/uploads/{artifactID}/download", h.RequestDownloadURL)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/art-1/download", nil)
	req = withAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// --- PresignRequest with DocumentSubtype tests ---

func TestUploadHandler_RequestPresignedURL_WithDocumentSubtype_Success(t *testing.T) {
	h := defaultUploadHandler()

	body, _ := json.Marshal(PresignRequest{
		FileName:        "inbody_scan.pdf",
		FileSize:        1024,
		ContentType:     "application/pdf",
		ClientID:        "client-1",
		DocumentSubtype: "inbody_pdf",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/presign", bytes.NewReader(body))
	req = withAuth(req)

	h.RequestPresignedURL(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// --- ConfirmUpload with DocumentSubtype tests ---

func TestUploadHandler_ConfirmUpload_WithDocumentSubtype_PassesHint(t *testing.T) {
	confirmMock := defaultConfirmUploadMock()
	h := NewUploadHandler(
		&mockRequestUploadUC{},
		confirmMock,
		&mockRequestDownloadUC{},
	)

	body, _ := json.Marshal(ConfirmRequest{
		DocumentSubtype: "inbody_pdf",
	})

	router := chi.NewRouter()
	router.Post("/api/v1/uploads/{artifactID}/confirm", h.ConfirmUpload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/art-1/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuth(req)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if confirmMock.lastInput.DocumentSubtype != entities.DocumentSubtypeInBodyPDF {
		t.Errorf("expected document_subtype %s, got %s", entities.DocumentSubtypeInBodyPDF, confirmMock.lastInput.DocumentSubtype)
	}
}

func TestUploadHandler_ConfirmUpload_WithoutBody_DefaultsEmpty(t *testing.T) {
	confirmMock := defaultConfirmUploadMock()
	h := NewUploadHandler(
		&mockRequestUploadUC{},
		confirmMock,
		&mockRequestDownloadUC{},
	)

	router := chi.NewRouter()
	router.Post("/api/v1/uploads/{artifactID}/confirm", h.ConfirmUpload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/art-1/confirm", nil)
	req = withAuth(req)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if confirmMock.lastInput.DocumentSubtype != "" {
		t.Errorf("expected empty document_subtype, got %s", confirmMock.lastInput.DocumentSubtype)
	}
}

func TestUploadHandler_ConfirmUpload_ResponseIncludesJobID(t *testing.T) {
	confirmMock := defaultConfirmUploadMock()
	h := NewUploadHandler(
		&mockRequestUploadUC{},
		confirmMock,
		&mockRequestDownloadUC{},
	)

	router := chi.NewRouter()
	router.Post("/api/v1/uploads/{artifactID}/confirm", h.ConfirmUpload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/uploads/art-1/confirm", nil)
	req = withAuth(req)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp usecase.ConfirmUploadOutput
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.JobID == "" {
		t.Error("expected non-empty job_id in response")
	}
	if resp.Artifact.DocumentSubtype != entities.DocumentSubtypeOther {
		t.Errorf("expected document_subtype %s, got %s", entities.DocumentSubtypeOther, resp.Artifact.DocumentSubtype)
	}
}
