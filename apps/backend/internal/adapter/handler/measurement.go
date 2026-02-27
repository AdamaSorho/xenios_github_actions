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

// MeasurementsUseCase defines the interface for retrieving client measurements.
type MeasurementsUseCase interface {
	Execute(ctx context.Context, input usecase.GetClientMeasurementsInput) (*entities.MeasurementPage, error)
}

// LatestMeasurementsUseCase defines the interface for retrieving latest measurements.
type LatestMeasurementsUseCase interface {
	Execute(ctx context.Context, coachID, clientID string) ([]*entities.Measurement, error)
}

// WearableSummariesUseCase defines the interface for retrieving wearable summaries.
type WearableSummariesUseCase interface {
	Execute(ctx context.Context, coachID, clientID string, limit int) ([]*entities.WearableSummary, error)
}

// ProfileSummaryUseCase defines the interface for retrieving profile summaries.
type ProfileSummaryUseCase interface {
	Execute(ctx context.Context, coachID, clientID string) (*entities.ProfileSummary, error)
}

// MeasurementHandler handles HTTP requests for client measurements and profile data.
type MeasurementHandler struct {
	measurementsUC MeasurementsUseCase
	latestUC       LatestMeasurementsUseCase
	wearableUC     WearableSummariesUseCase
	profileUC      ProfileSummaryUseCase
}

// NewMeasurementHandler creates a new MeasurementHandler.
func NewMeasurementHandler(
	measurementsUC MeasurementsUseCase,
	latestUC LatestMeasurementsUseCase,
	wearableUC WearableSummariesUseCase,
	profileUC ProfileSummaryUseCase,
) *MeasurementHandler {
	return &MeasurementHandler{
		measurementsUC: measurementsUC,
		latestUC:       latestUC,
		wearableUC:     wearableUC,
		profileUC:      profileUC,
	}
}

// ListMeasurements handles GET /api/v1/clients/{clientID}/measurements
func (h *MeasurementHandler) ListMeasurements(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing client ID", "INVALID_REQUEST", nil)
		return
	}

	coachID := extractCoachID(r)
	if coachID == "" {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	filter := parseMeasurementFilter(r)

	result, err := h.measurementsUC.Execute(r.Context(), usecase.GetClientMeasurementsInput{
		CoachID:  coachID,
		ClientID: clientID,
		Filter:   filter,
	})
	if err != nil {
		handleMeasurementError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, result)
}

// LatestMeasurements handles GET /api/v1/clients/{clientID}/measurements/latest
func (h *MeasurementHandler) LatestMeasurements(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing client ID", "INVALID_REQUEST", nil)
		return
	}

	coachID := extractCoachID(r)
	if coachID == "" {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	result, err := h.latestUC.Execute(r.Context(), coachID, clientID)
	if err != nil {
		handleMeasurementError(w, err)
		return
	}

	if result == nil {
		result = []*entities.Measurement{}
	}

	_ = respondJSON(w, http.StatusOK, map[string]interface{}{"measurements": result})
}

// WearableSummaries handles GET /api/v1/clients/{clientID}/wearable-summaries
func (h *MeasurementHandler) WearableSummaries(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing client ID", "INVALID_REQUEST", nil)
		return
	}

	coachID := extractCoachID(r)
	if coachID == "" {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	result, err := h.wearableUC.Execute(r.Context(), coachID, clientID, limit)
	if err != nil {
		handleMeasurementError(w, err)
		return
	}

	if result == nil {
		result = []*entities.WearableSummary{}
	}

	_ = respondJSON(w, http.StatusOK, map[string]interface{}{"summaries": result})
}

// ProfileSummary handles GET /api/v1/clients/{clientID}/profile-summary
func (h *MeasurementHandler) ProfileSummary(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing client ID", "INVALID_REQUEST", nil)
		return
	}

	coachID := extractCoachID(r)
	if coachID == "" {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	result, err := h.profileUC.Execute(r.Context(), coachID, clientID)
	if err != nil {
		handleMeasurementError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, result)
}

func extractCoachID(r *http.Request) string {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		return ""
	}
	return claims.Subject
}

func parseMeasurementFilter(r *http.Request) entities.MeasurementFilter {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := entities.MeasurementFilter{
		Type:  q.Get("type"),
		Page:  page,
		Limit: limit,
	}

	if from := q.Get("from"); from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			filter.From = &t
		}
	}
	if to := q.Get("to"); to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			filter.To = &t
		}
	}

	return filter
}

func handleMeasurementError(w http.ResponseWriter, err error) {
	if usecase.IsValidationError(err) {
		respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
		return
	}
	if usecase.IsAuthorizationError(err) {
		respondErrorWithCode(w, http.StatusForbidden, "forbidden", "FORBIDDEN", nil)
		return
	}
	respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
}
