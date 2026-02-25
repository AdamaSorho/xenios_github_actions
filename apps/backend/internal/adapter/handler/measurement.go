package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/xenios/backend/internal/adapter/middleware"
	"github.com/xenios/backend/internal/usecase"
)

// GetClientMeasurementsUseCase defines the interface for retrieving client measurements.
type GetClientMeasurementsUseCase interface {
	Execute(ctx context.Context, input usecase.GetClientMeasurementsInput) (*usecase.GetClientMeasurementsOutput, error)
}

// GetLatestMeasurementsUseCase defines the interface for retrieving latest measurements.
type GetLatestMeasurementsUseCase interface {
	Execute(ctx context.Context, input usecase.GetLatestMeasurementsInput) (*usecase.GetLatestMeasurementsOutput, error)
}

// GetWearableSummariesUseCase defines the interface for retrieving wearable summaries.
type GetWearableSummariesUseCase interface {
	Execute(ctx context.Context, input usecase.GetWearableSummariesInput) (*usecase.GetWearableSummariesOutput, error)
}

// GetClientProfileSummaryUseCase defines the interface for retrieving client profile summary.
type GetClientProfileSummaryUseCase interface {
	Execute(ctx context.Context, input usecase.GetClientProfileSummaryInput) (*usecase.ProfileSummaryOutput, error)
}

// MeasurementHandler handles HTTP requests for measurement and profile endpoints.
type MeasurementHandler struct {
	getMeasurementsUC      GetClientMeasurementsUseCase
	getLatestUC            GetLatestMeasurementsUseCase
	getWearableSummariesUC GetWearableSummariesUseCase
	getProfileSummaryUC    GetClientProfileSummaryUseCase
}

// NewMeasurementHandler creates a new MeasurementHandler.
func NewMeasurementHandler(
	getMeasurementsUC GetClientMeasurementsUseCase,
	getLatestUC GetLatestMeasurementsUseCase,
	getWearableSummariesUC GetWearableSummariesUseCase,
	getProfileSummaryUC GetClientProfileSummaryUseCase,
) *MeasurementHandler {
	return &MeasurementHandler{
		getMeasurementsUC:      getMeasurementsUC,
		getLatestUC:            getLatestUC,
		getWearableSummariesUC: getWearableSummariesUC,
		getProfileSummaryUC:    getProfileSummaryUC,
	}
}

// ListMeasurements handles GET /api/v1/clients/{clientID}/measurements
func (h *MeasurementHandler) ListMeasurements(w http.ResponseWriter, r *http.Request) {
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

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	input := usecase.GetClientMeasurementsInput{
		CoachID:  claims.Subject,
		ClientID: clientID,
		Type:     q.Get("type"),
		Page:     page,
		Limit:    limit,
	}

	if fromStr := q.Get("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			input.From = &t
		} else if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			input.From = &t
		}
	}
	if toStr := q.Get("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			input.To = &t
		} else if t, err := time.Parse("2006-01-02", toStr); err == nil {
			input.To = &t
		}
	}

	out, err := h.getMeasurementsUC.Execute(r.Context(), input)
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
}

// LatestMeasurements handles GET /api/v1/clients/{clientID}/measurements/latest
func (h *MeasurementHandler) LatestMeasurements(w http.ResponseWriter, r *http.Request) {
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

	out, err := h.getLatestUC.Execute(r.Context(), usecase.GetLatestMeasurementsInput{
		CoachID:  claims.Subject,
		ClientID: clientID,
	})
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
}

// WearableSummaries handles GET /api/v1/clients/{clientID}/wearable-summaries
func (h *MeasurementHandler) WearableSummaries(w http.ResponseWriter, r *http.Request) {
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

	days, _ := strconv.Atoi(r.URL.Query().Get("days"))

	out, err := h.getWearableSummariesUC.Execute(r.Context(), usecase.GetWearableSummariesInput{
		CoachID:  claims.Subject,
		ClientID: clientID,
		Days:     days,
	})
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
}

// ProfileSummary handles GET /api/v1/clients/{clientID}/profile-summary
func (h *MeasurementHandler) ProfileSummary(w http.ResponseWriter, r *http.Request) {
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

	out, err := h.getProfileSummaryUC.Execute(r.Context(), usecase.GetClientProfileSummaryInput{
		CoachID:  claims.Subject,
		ClientID: clientID,
	})
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
}

func handleUseCaseError(w http.ResponseWriter, err error) {
	if usecase.IsValidationError(err) {
		respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
		return
	}
	if usecase.IsAuthorizationError(err) {
		respondErrorWithCode(w, http.StatusForbidden, "access denied", "FORBIDDEN", nil)
		return
	}
	respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
}
