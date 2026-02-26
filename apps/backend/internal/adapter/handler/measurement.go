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

// GetClientMeasurementsUC defines the interface for the get client measurements use case.
type GetClientMeasurementsUC interface {
	Execute(ctx context.Context, input usecase.GetClientMeasurementsInput) (*entities.MeasurementResult, error)
}

// GetLatestMeasurementsUC defines the interface for the get latest measurements use case.
type GetLatestMeasurementsUC interface {
	Execute(ctx context.Context, input usecase.GetLatestMeasurementsInput) ([]*entities.LatestMeasurement, error)
}

// GetWearableSummariesUC defines the interface for the get wearable summaries use case.
type GetWearableSummariesUC interface {
	Execute(ctx context.Context, input usecase.GetWearableSummariesInput) ([]*entities.WearableSummary, error)
}

// GetClientProfileSummaryUC defines the interface for the get client profile summary use case.
type GetClientProfileSummaryUC interface {
	Execute(ctx context.Context, input usecase.GetClientProfileSummaryInput) (*entities.ProfileSummary, error)
}

// MeasurementHandler handles HTTP requests for client measurements and profile data.
type MeasurementHandler struct {
	getMeasurementsUC    GetClientMeasurementsUC
	getLatestUC          GetLatestMeasurementsUC
	getWearableSumUC     GetWearableSummariesUC
	getProfileSummaryUC  GetClientProfileSummaryUC
}

// NewMeasurementHandler creates a new MeasurementHandler.
func NewMeasurementHandler(
	getMeasurementsUC GetClientMeasurementsUC,
	getLatestUC GetLatestMeasurementsUC,
	getWearableSumUC GetWearableSummariesUC,
	getProfileSummaryUC GetClientProfileSummaryUC,
) *MeasurementHandler {
	return &MeasurementHandler{
		getMeasurementsUC:   getMeasurementsUC,
		getLatestUC:         getLatestUC,
		getWearableSumUC:    getWearableSumUC,
		getProfileSummaryUC: getProfileSummaryUC,
	}
}

// ListMeasurements handles GET /api/v1/clients/{clientID}/measurements
func (h *MeasurementHandler) ListMeasurements(w http.ResponseWriter, r *http.Request) {
	coachID, ok := extractCoachID(w, r)
	if !ok {
		return
	}
	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing client ID", "INVALID_REQUEST", nil)
		return
	}

	filter := parseMeasurementFilter(r)

	result, err := h.getMeasurementsUC.Execute(r.Context(), usecase.GetClientMeasurementsInput{
		CoachID:  coachID,
		ClientID: clientID,
		Filter:   filter,
	})
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, result)
}

// LatestMeasurements handles GET /api/v1/clients/{clientID}/measurements/latest
func (h *MeasurementHandler) LatestMeasurements(w http.ResponseWriter, r *http.Request) {
	coachID, ok := extractCoachID(w, r)
	if !ok {
		return
	}
	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing client ID", "INVALID_REQUEST", nil)
		return
	}

	results, err := h.getLatestUC.Execute(r.Context(), usecase.GetLatestMeasurementsInput{
		CoachID:  coachID,
		ClientID: clientID,
	})
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, map[string]interface{}{"measurements": results})
}

// WearableSummaries handles GET /api/v1/clients/{clientID}/wearable-summaries
func (h *MeasurementHandler) WearableSummaries(w http.ResponseWriter, r *http.Request) {
	coachID, ok := extractCoachID(w, r)
	if !ok {
		return
	}
	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing client ID", "INVALID_REQUEST", nil)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	results, err := h.getWearableSumUC.Execute(r.Context(), usecase.GetWearableSummariesInput{
		CoachID:  coachID,
		ClientID: clientID,
		Limit:    limit,
	})
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, map[string]interface{}{"summaries": results})
}

// ProfileSummary handles GET /api/v1/clients/{clientID}/profile-summary
func (h *MeasurementHandler) ProfileSummary(w http.ResponseWriter, r *http.Request) {
	coachID, ok := extractCoachID(w, r)
	if !ok {
		return
	}
	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing client ID", "INVALID_REQUEST", nil)
		return
	}

	result, err := h.getProfileSummaryUC.Execute(r.Context(), usecase.GetClientProfileSummaryInput{
		CoachID:  coachID,
		ClientID: clientID,
	})
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, result)
}

func extractCoachID(w http.ResponseWriter, r *http.Request) (string, bool) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return "", false
	}
	return claims.Subject, true
}

func parseMeasurementFilter(r *http.Request) entities.MeasurementFilter {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := entities.MeasurementFilter{
		MeasurementType: q.Get("type"),
		Page:            page,
		Limit:           limit,
	}

	if fromStr := q.Get("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filter.From = &t
		} else if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			filter.From = &t
		}
	}
	if toStr := q.Get("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			filter.To = &t
		} else if t, err := time.Parse("2006-01-02", toStr); err == nil {
			filter.To = &t
		}
	}

	return filter
}

func handleUseCaseError(w http.ResponseWriter, err error) {
	if usecase.IsValidationError(err) {
		respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
		return
	}
	if usecase.IsAuthorizationError(err) {
		respondErrorWithCode(w, http.StatusForbidden, err.Error(), "FORBIDDEN", nil)
		return
	}
	if usecase.IsAuthenticationError(err) {
		respondErrorWithCode(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED", nil)
		return
	}
	respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
}
