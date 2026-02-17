package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/xenios/backend/internal/adapter/middleware"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// RequestUploadUseCase defines the interface for the request upload use case.
type RequestUploadUseCase interface {
	Execute(ctx context.Context, input usecase.RequestUploadInput) (*usecase.RequestUploadOutput, error)
}

// ConfirmUploadUseCase defines the interface for the confirm upload use case.
type ConfirmUploadUseCase interface {
	Execute(ctx context.Context, input usecase.ConfirmUploadInput) (*usecase.ConfirmUploadOutput, error)
}

// RequestDownloadUseCase defines the interface for the request download use case.
type RequestDownloadUseCase interface {
	Execute(ctx context.Context, input usecase.RequestDownloadInput) (*usecase.RequestDownloadOutput, error)
}

// UploadHandler handles HTTP requests for file uploads and downloads.
type UploadHandler struct {
	requestUploadUC  RequestUploadUseCase
	confirmUploadUC  ConfirmUploadUseCase
	requestDownloadUC RequestDownloadUseCase
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(
	requestUploadUC RequestUploadUseCase,
	confirmUploadUC ConfirmUploadUseCase,
	requestDownloadUC RequestDownloadUseCase,
) *UploadHandler {
	return &UploadHandler{
		requestUploadUC:   requestUploadUC,
		confirmUploadUC:   confirmUploadUC,
		requestDownloadUC: requestDownloadUC,
	}
}

// PresignRequest is the JSON request body for requesting a presigned upload URL.
type PresignRequest struct {
	FileName        string `json:"file_name"`
	FileSize        int64  `json:"file_size"`
	ContentType     string `json:"content_type"`
	ClientID        string `json:"client_id"`
	DocumentSubtype string `json:"document_subtype,omitempty"`
}

// RequestPresignedURL handles POST /api/v1/uploads/presign
func (h *UploadHandler) RequestPresignedURL(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	var req PresignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErrorWithCode(w, http.StatusBadRequest, "invalid JSON body", "INVALID_JSON", nil)
		return
	}

	out, err := h.requestUploadUC.Execute(r.Context(), usecase.RequestUploadInput{
		FileName:        req.FileName,
		FileSize:        req.FileSize,
		ContentType:     req.ContentType,
		ClientID:        req.ClientID,
		CoachID:         claims.Subject,
		DocumentSubtype: entities.DocumentSubtype(req.DocumentSubtype),
	})
	if err != nil {
		if usecase.IsValidationError(err) {
			respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
			return
		}
		respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
}

// ConfirmUpload handles POST /api/v1/uploads/{artifactID}/confirm
func (h *UploadHandler) ConfirmUpload(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	artifactID := chi.URLParam(r, "artifactID")
	if artifactID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing artifact ID", "INVALID_REQUEST", nil)
		return
	}

	out, err := h.confirmUploadUC.Execute(r.Context(), usecase.ConfirmUploadInput{
		ArtifactID: artifactID,
		CoachID:    claims.Subject,
	})
	if err != nil {
		if usecase.IsValidationError(err) {
			respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
			return
		}
		if usecase.IsAuthenticationError(err) {
			respondErrorWithCode(w, http.StatusForbidden, err.Error(), "FORBIDDEN", nil)
			return
		}
		respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
}

// RequestDownloadURL handles POST /api/v1/uploads/{artifactID}/download
func (h *UploadHandler) RequestDownloadURL(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	artifactID := chi.URLParam(r, "artifactID")
	if artifactID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing artifact ID", "INVALID_REQUEST", nil)
		return
	}

	out, err := h.requestDownloadUC.Execute(r.Context(), usecase.RequestDownloadInput{
		ArtifactID: artifactID,
		CoachID:    claims.Subject,
	})
	if err != nil {
		if usecase.IsValidationError(err) {
			respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
			return
		}
		if usecase.IsAuthenticationError(err) {
			respondErrorWithCode(w, http.StatusForbidden, err.Error(), "FORBIDDEN", nil)
			return
		}
		respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
}
