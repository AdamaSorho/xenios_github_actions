package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// GenerateInsightsUseCase defines the interface for insight generation.
type GenerateInsightsUseCase interface {
	Execute(ctx context.Context, input usecase.GenerateInsightsInput) (*usecase.GenerateInsightsOutput, error)
}

// ListInsightsUseCase defines the interface for listing insights by status.
type ListInsightsUseCase interface {
	Execute(ctx context.Context, coachID string, status entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, error)
}

// UpdateInsightStatusUseCase defines the interface for updating insight status.
type UpdateInsightStatusUseCase interface {
	Execute(ctx context.Context, id string, status entities.InsightStatus) (*entities.InsightCard, error)
}

// InsightHandler handles HTTP requests for insight card operations.
type InsightHandler struct {
	listUseCase         ListInsightsUseCase
	updateStatusUseCase UpdateInsightStatusUseCase
}

// NewInsightHandler creates a new InsightHandler.
func NewInsightHandler(listUC ListInsightsUseCase, updateStatusUC UpdateInsightStatusUseCase) *InsightHandler {
	return &InsightHandler{
		listUseCase:         listUC,
		updateStatusUseCase: updateStatusUC,
	}
}

// ListDraftInsights handles GET /coaches/{coachID}/insights/drafts.
func (h *InsightHandler) ListDraftInsights(w http.ResponseWriter, r *http.Request) {
	coachID := chi.URLParam(r, "coachID")
	if coachID == "" {
		respondError(w, http.StatusBadRequest, "coach_id is required")
		return
	}

	cards, err := h.listUseCase.Execute(r.Context(), coachID, entities.InsightStatusDraft, 50, 0)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list insights")
		return
	}

	if cards == nil {
		cards = []*entities.InsightCard{}
	}

	_ = respondJSON(w, http.StatusOK, cards)
}

// UpdateInsightStatusRequest is the JSON request body for updating insight status.
type UpdateInsightStatusRequest struct {
	Status string `json:"status"`
}

// UpdateInsightStatus handles PUT /insights/{insightID}/status.
func (h *InsightHandler) UpdateInsightStatus(w http.ResponseWriter, r *http.Request) {
	insightID := chi.URLParam(r, "insightID")
	if insightID == "" {
		respondError(w, http.StatusBadRequest, "insight_id is required")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)

	var req UpdateInsightStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Status == "" {
		respondError(w, http.StatusBadRequest, "status is required")
		return
	}

	status := entities.InsightStatus(req.Status)
	if !entities.IsValidInsightStatus(status) {
		respondError(w, http.StatusBadRequest, "invalid status value")
		return
	}

	card, err := h.updateStatusUseCase.Execute(r.Context(), insightID, status)
	if err != nil {
		var validationErr *usecase.ValidationError
		if errors.As(err, &validationErr) {
			respondError(w, http.StatusBadRequest, validationErr.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update insight status")
		return
	}

	_ = respondJSON(w, http.StatusOK, card)
}
