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

// GetClientMeasurementsUseCase defines the interface for the get client measurements use case.
type GetClientMeasurementsUseCase interface {
	Execute(ctx context.Context, input usecase.GetMeasurementsInput) (*usecase.GetMeasurementsOutput, error)
}

// GetLatestMeasurementsUseCase defines the interface for the get latest measurements use case.
type GetLatestMeasurementsUseCase interface {
	Execute(ctx context.Context, coachID, clientID string) ([]*entities.LatestMeasurement, error)
}

// GetWearableSummariesUseCase defines the interface for the get wearable summaries use case.
type GetWearableSummariesUseCase interface {
	Execute(ctx context.Context, input usecase.GetWearableSummariesInput) (*usecase.GetWearableSummariesOutput, error)
}

// GetClientProfileSummaryUseCase defines the interface for the get client profile summary use case.
type GetClientProfileSummaryUseCase interface {
	Execute(ctx context.Context, coachID, clientID string) (*entities.ProfileSummary, error)
}

// MeasurementHandler handles HTTP requests for client measurements and profile data.
type MeasurementHandler struct {
	getMeasurementsUC    GetClientMeasurementsUseCase
	getLatestUC          GetLatestMeasurementsUseCase
	getWearableSummaries GetWearableSummariesUseCase
	getProfileSummaryUC  GetClientProfileSummaryUseCase
}

// NewMeasurementHandler creates a new MeasurementHandler.
func NewMeasurementHandler(
	getMeasurementsUC GetClientMeasurementsUseCase,
	getLatestUC GetLatestMeasurementsUseCase,
	getWearableSummaries GetWearableSummariesUseCase,
	getProfileSummaryUC GetClientProfileSummaryUseCase,
) *MeasurementHandler {
	return &MeasurementHandler{
		getMeasurementsUC:    getMeasurementsUC,
		getLatestUC:          getLatestUC,
		getWearableSummaries: getWearableSummaries,
		getProfileSummaryUC:  getProfileSummaryUC,
	}
}

// ListMeasurements handles GET /api/v1/clients/{clientID}/measurements
func (h *MeasurementHandler) ListMeasurements(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing client ID", "INVALID_REQUEST", nil)
		return
	}

	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	measurementType := r.URL.Query().Get("type")

	var from, to *time.Time
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		parsed, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			parsed, err = time.Parse("2006-01-02", fromStr)
			if err != nil {
				respondErrorWithCode(w, http.StatusBadRequest, "invalid 'from' date format", "VALIDATION_ERROR", nil)
				return
			}
		}
		from = &parsed
	}
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		parsed, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			parsed, err = time.Parse("2006-01-02", toStr)
			if err != nil {
				respondErrorWithCode(w, http.StatusBadRequest, "invalid 'to' date format", "VALIDATION_ERROR", nil)
				return
			}
		}
		to = &parsed
	}

	input := usecase.GetMeasurementsInput{
		CoachID:         claims.Subject,
		ClientID:        clientID,
		MeasurementType: measurementType,
		From:            from,
		To:              to,
		Page:            page,
		Limit:           limit,
	}

	output, err := h.getMeasurementsUC.Execute(r.Context(), input)
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, output)
}

// LatestMeasurements handles GET /api/v1/clients/{clientID}/measurements/latest
func (h *MeasurementHandler) LatestMeasurements(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing client ID", "INVALID_REQUEST", nil)
		return
	}

	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	results, err := h.getLatestUC.Execute(r.Context(), claims.Subject, clientID)
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, map[string]interface{}{
		"measurements": results,
	})
}

// ListWearableSummaries handles GET /api/v1/clients/{clientID}/wearable-summaries
func (h *MeasurementHandler) ListWearableSummaries(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing client ID", "INVALID_REQUEST", nil)
		return
	}

	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	source := r.URL.Query().Get("source")

	var from, to *time.Time
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		parsed, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			respondErrorWithCode(w, http.StatusBadRequest, "invalid 'from' date format", "VALIDATION_ERROR", nil)
			return
		}
		from = &parsed
	}
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		parsed, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			respondErrorWithCode(w, http.StatusBadRequest, "invalid 'to' date format", "VALIDATION_ERROR", nil)
			return
		}
		to = &parsed
	}

	input := usecase.GetWearableSummariesInput{
		CoachID:  claims.Subject,
		ClientID: clientID,
		Source:   source,
		From:     from,
		To:       to,
		Page:     page,
		Limit:    limit,
	}

	output, err := h.getWearableSummaries.Execute(r.Context(), input)
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, output)
}

// ProfileSummary handles GET /api/v1/clients/{clientID}/profile-summary
func (h *MeasurementHandler) ProfileSummary(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing client ID", "INVALID_REQUEST", nil)
		return
	}

	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	summary, err := h.getProfileSummaryUC.Execute(r.Context(), claims.Subject, clientID)
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, summary)
}

// handleUseCaseError maps use case errors to appropriate HTTP responses.
func handleUseCaseError(w http.ResponseWriter, err error) {
	if usecase.IsValidationError(err) {
		respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
		return
	}
	if usecase.IsAuthorizationError(err) {
		respondErrorWithCode(w, http.StatusForbidden, "access denied", "FORBIDDEN", nil)
		return
	}
	if usecase.IsAuthenticationError(err) {
		respondErrorWithCode(w, http.StatusUnauthorized, "unauthorized", "UNAUTHORIZED", nil)
		return
	}
	respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
}
