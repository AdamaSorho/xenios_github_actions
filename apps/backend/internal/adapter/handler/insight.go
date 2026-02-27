package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/xenios/backend/internal/adapter/middleware"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// InsightTransitionUseCase defines the interface for insight status transitions.
type InsightTransitionUseCase interface {
	Execute(ctx context.Context, input usecase.TransitionInput) (*entities.InsightCard, error)
}

// EditInsightUseCase defines the interface for editing an insight card.
type EditInsightUseCase interface {
	Execute(ctx context.Context, input usecase.EditInsightInput) (*entities.InsightCard, error)
}

// GetInsightQueueUseCase defines the interface for retrieving the insight queue.
type GetInsightQueueUseCase interface {
	Execute(ctx context.Context, input usecase.InsightQueueInput) (*usecase.InsightQueueOutput, error)
}

// GetClientInsightsUseCase defines the interface for retrieving client insights.
type GetClientInsightsUseCase interface {
	Execute(ctx context.Context, input usecase.ClientInsightsInput) (*usecase.InsightQueueOutput, error)
}

// InsightHandler handles HTTP requests for insight card operations.
type InsightHandler struct {
	transitionUC     InsightTransitionUseCase
	editUC           EditInsightUseCase
	getQueueUC       GetInsightQueueUseCase
	getClientUC      GetClientInsightsUseCase
}

// NewInsightHandler creates a new InsightHandler.
func NewInsightHandler(
	transitionUC InsightTransitionUseCase,
	editUC EditInsightUseCase,
	getQueueUC GetInsightQueueUseCase,
	getClientUC GetClientInsightsUseCase,
) *InsightHandler {
	return &InsightHandler{
		transitionUC: transitionUC,
		editUC:       editUC,
		getQueueUC:   getQueueUC,
		getClientUC:  getClientUC,
	}
}

// GetQueue handles GET /api/v1/insights/queue
func (h *InsightHandler) GetQueue(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	page, limit := parsePagination(r)

	out, err := h.getQueueUC.Execute(r.Context(), usecase.InsightQueueInput{
		CoachID: claims.Subject,
		Page:    page,
		Limit:   limit,
	})
	if err != nil {
		handleInsightError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
}

// GetClientInsights handles GET /api/v1/clients/{clientID}/insights
func (h *InsightHandler) GetClientInsights(w http.ResponseWriter, r *http.Request) {
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

	page, limit := parsePagination(r)
	var status *entities.InsightStatus
	if s := r.URL.Query().Get("status"); s != "" {
		is := entities.InsightStatus(s)
		if !entities.ValidInsightStatuses[is] {
			respondErrorWithCode(w, http.StatusBadRequest, "invalid status filter", "INVALID_REQUEST", nil)
			return
		}
		status = &is
	}

	out, err := h.getClientUC.Execute(r.Context(), usecase.ClientInsightsInput{
		ClientID: clientID,
		CoachID:  claims.Subject,
		Status:   status,
		Page:     page,
		Limit:    limit,
	})
	if err != nil {
		handleInsightError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
}

// editInsightRequest is the JSON request body for editing an insight.
type editInsightRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// EditInsight handles PUT /api/v1/insights/{insightID}
func (h *InsightHandler) EditInsight(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	insightID := chi.URLParam(r, "insightID")
	if insightID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing insight ID", "INVALID_REQUEST", nil)
		return
	}

	var req editInsightRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErrorWithCode(w, http.StatusBadRequest, "invalid JSON body", "INVALID_JSON", nil)
		return
	}

	out, err := h.editUC.Execute(r.Context(), usecase.EditInsightInput{
		InsightID: insightID,
		CoachID:   claims.Subject,
		Title:     req.Title,
		Body:      req.Body,
	})
	if err != nil {
		handleInsightError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
}

// Approve handles PUT /api/v1/insights/{insightID}/approve
func (h *InsightHandler) Approve(w http.ResponseWriter, r *http.Request) {
	h.handleTransition(w, r, entities.InsightStatusApproved, "insight.approved")
}

// Dismiss handles PUT /api/v1/insights/{insightID}/dismiss
func (h *InsightHandler) Dismiss(w http.ResponseWriter, r *http.Request) {
	h.handleTransition(w, r, entities.InsightStatusDismissed, "insight.dismissed")
}

// Share handles PUT /api/v1/insights/{insightID}/share
func (h *InsightHandler) Share(w http.ResponseWriter, r *http.Request) {
	h.handleTransition(w, r, entities.InsightStatusShared, "insight.shared")
}

// handleTransition is a shared handler for status transition endpoints.
func (h *InsightHandler) handleTransition(w http.ResponseWriter, r *http.Request, target entities.InsightStatus, auditAction string) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	insightID := chi.URLParam(r, "insightID")
	if insightID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing insight ID", "INVALID_REQUEST", nil)
		return
	}

	out, err := h.transitionUC.Execute(r.Context(), usecase.TransitionInput{
		InsightID:    insightID,
		CoachID:      claims.Subject,
		TargetStatus: target,
		AuditAction:  auditAction,
	})
	if err != nil {
		handleInsightError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
}

// handleInsightError maps use case errors to HTTP responses.
func handleInsightError(w http.ResponseWriter, err error) {
	if usecase.IsValidationError(err) {
		respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
		return
	}
	if usecase.IsAuthenticationError(err) {
		respondErrorWithCode(w, http.StatusForbidden, err.Error(), "FORBIDDEN", nil)
		return
	}
	if usecase.IsTransitionError(err) {
		respondErrorWithCode(w, http.StatusUnprocessableEntity, err.Error(), "INVALID_TRANSITION", nil)
		return
	}
	respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
}

// parsePagination extracts page and limit from query parameters.
func parsePagination(r *http.Request) (int, int) {
	page := 1
	limit := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	return page, limit
}
