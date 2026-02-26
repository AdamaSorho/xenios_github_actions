package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/xenios/backend/internal/adapter/middleware"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// GetClientMeasurementsUC defines the interface for the get measurements use case.
type GetClientMeasurementsUC interface {
	Execute(ctx context.Context, coachID string, filter entities.MeasurementFilter) (*usecase.GetClientMeasurementsOutput, error)
}

// GetLatestMeasurementsUC defines the interface for the get latest measurements use case.
type GetLatestMeasurementsUC interface {
	Execute(ctx context.Context, coachID, clientID string) (*usecase.GetLatestMeasurementsOutput, error)
}

// GetWearableSummariesUC defines the interface for the get wearable summaries use case.
type GetWearableSummariesUC interface {
	Execute(ctx context.Context, coachID, clientID string, limit, offset int) (*usecase.GetWearableSummariesOutput, error)
}

// GetClientProfileSummaryUC defines the interface for the profile summary use case.
type GetClientProfileSummaryUC interface {
	Execute(ctx context.Context, coachID, clientID string) (*usecase.ProfileSummaryOutput, error)
}

// MeasurementHandler handles HTTP requests for client measurements and profile data.
type MeasurementHandler struct {
	getMeasurementsUC    GetClientMeasurementsUC
	getLatestUC          GetLatestMeasurementsUC
	getWearablesUC       GetWearableSummariesUC
	getProfileSummaryUC  GetClientProfileSummaryUC
}

// NewMeasurementHandler creates a new MeasurementHandler.
func NewMeasurementHandler(
	getMeasurementsUC GetClientMeasurementsUC,
	getLatestUC GetLatestMeasurementsUC,
	getWearablesUC GetWearableSummariesUC,
	getProfileSummaryUC GetClientProfileSummaryUC,
) *MeasurementHandler {
	return &MeasurementHandler{
		getMeasurementsUC:   getMeasurementsUC,
		getLatestUC:         getLatestUC,
		getWearablesUC:      getWearablesUC,
		getProfileSummaryUC: getProfileSummaryUC,
	}
}

// GetMeasurements handles GET /api/v1/clients/{clientID}/measurements
func (h *MeasurementHandler) GetMeasurements(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing client ID", "INVALID_REQUEST", nil)
		return
	}

	filter := entities.MeasurementFilter{
		ClientID:        clientID,
		MeasurementType: r.URL.Query().Get("type"),
	}

	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filter.From = &t
		} else if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			filter.From = &t
		}
	}
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			filter.To = &t
		} else if t, err := time.Parse("2006-01-02", toStr); err == nil {
			filter.To = &t
		}
	}

	filter.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	filter.Offset = pageToOffset(r.URL.Query().Get("page"), filter.Limit)

	out, err := h.getMeasurementsUC.Execute(r.Context(), claims.Subject, filter)
	if err != nil {
		handleMeasurementError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
}

// GetLatestMeasurements handles GET /api/v1/clients/{clientID}/measurements/latest
func (h *MeasurementHandler) GetLatestMeasurements(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing client ID", "INVALID_REQUEST", nil)
		return
	}

	out, err := h.getLatestUC.Execute(r.Context(), claims.Subject, clientID)
	if err != nil {
		handleMeasurementError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
}

// GetWearableSummaries handles GET /api/v1/clients/{clientID}/wearable-summaries
func (h *MeasurementHandler) GetWearableSummaries(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing client ID", "INVALID_REQUEST", nil)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	out, err := h.getWearablesUC.Execute(r.Context(), claims.Subject, clientID, limit, offset)
	if err != nil {
		handleMeasurementError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
}

// GetProfileSummary handles GET /api/v1/clients/{clientID}/profile-summary
func (h *MeasurementHandler) GetProfileSummary(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing client ID", "INVALID_REQUEST", nil)
		return
	}

	out, err := h.getProfileSummaryUC.Execute(r.Context(), claims.Subject, clientID)
	if err != nil {
		handleMeasurementError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
}

func handleMeasurementError(w http.ResponseWriter, err error) {
	if usecase.IsValidationError(err) {
		respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
		return
	}
	if usecase.IsAuthenticationError(err) {
		respondErrorWithCode(w, http.StatusForbidden, err.Error(), "FORBIDDEN", nil)
		return
	}
	respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
}

func pageToOffset(pageStr string, limit int) int {
	page, _ := strconv.Atoi(pageStr)
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	return (page - 1) * limit
}
