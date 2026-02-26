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

// InsightQueueUseCase defines the interface for the get insight queue use case.
type InsightQueueUseCase interface {
	Execute(ctx context.Context, input usecase.GetInsightQueueInput) (*usecase.GetInsightQueueOutput, error)
}

// ClientInsightsUseCase defines the interface for the get client insights use case.
type ClientInsightsUseCase interface {
	Execute(ctx context.Context, input usecase.GetClientInsightsInput) (*usecase.GetInsightQueueOutput, error)
}

// ApproveInsightUseCaseIface defines the interface for approving insights.
type ApproveInsightUseCaseIface interface {
	Execute(ctx context.Context, input usecase.InsightActionInput) (*entities.InsightCard, error)
}

// DismissInsightUseCaseIface defines the interface for dismissing insights.
type DismissInsightUseCaseIface interface {
	Execute(ctx context.Context, input usecase.InsightActionInput) (*entities.InsightCard, error)
}

// EditInsightUseCaseIface defines the interface for editing insights.
type EditInsightUseCaseIface interface {
	Execute(ctx context.Context, input usecase.EditInsightInput) (*entities.InsightCard, error)
}

// ShareInsightUseCaseIface defines the interface for sharing insights.
type ShareInsightUseCaseIface interface {
	Execute(ctx context.Context, input usecase.InsightActionInput) (*entities.InsightCard, error)
}

// InsightHandler handles HTTP requests for insight card management.
type InsightHandler struct {
	queueUC          InsightQueueUseCase
	clientInsightsUC ClientInsightsUseCase
	approveUC        ApproveInsightUseCaseIface
	dismissUC        DismissInsightUseCaseIface
	editUC           EditInsightUseCaseIface
	shareUC          ShareInsightUseCaseIface
}

// NewInsightHandler creates a new InsightHandler.
func NewInsightHandler(
	queueUC InsightQueueUseCase,
	clientInsightsUC ClientInsightsUseCase,
	approveUC ApproveInsightUseCaseIface,
	dismissUC DismissInsightUseCaseIface,
	editUC EditInsightUseCaseIface,
	shareUC ShareInsightUseCaseIface,
) *InsightHandler {
	return &InsightHandler{
		queueUC:          queueUC,
		clientInsightsUC: clientInsightsUC,
		approveUC:        approveUC,
		dismissUC:        dismissUC,
		editUC:           editUC,
		shareUC:          shareUC,
	}
}

// GetQueue handles GET /api/v1/insights/queue
func (h *InsightHandler) GetQueue(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	status := r.URL.Query().Get("status")

	out, err := h.queueUC.Execute(r.Context(), usecase.GetInsightQueueInput{
		CoachID: claims.Subject,
		Status:  status,
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

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	status := r.URL.Query().Get("status")

	out, err := h.clientInsightsUC.Execute(r.Context(), usecase.GetClientInsightsInput{
		CoachID:  claims.Subject,
		ClientID: clientID,
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

// Approve handles PUT /api/v1/insights/{insightID}/approve
func (h *InsightHandler) Approve(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.approveUC.Execute(r.Context(), usecase.InsightActionInput{
		InsightID: insightID,
		CoachID:   claims.Subject,
	})
	if err != nil {
		handleInsightError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, result)
}

// Dismiss handles PUT /api/v1/insights/{insightID}/dismiss
func (h *InsightHandler) Dismiss(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.dismissUC.Execute(r.Context(), usecase.InsightActionInput{
		InsightID: insightID,
		CoachID:   claims.Subject,
	})
	if err != nil {
		handleInsightError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, result)
}

// EditInsightRequest is the JSON request body for editing an insight.
type EditInsightRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Edit handles PUT /api/v1/insights/{insightID}
func (h *InsightHandler) Edit(w http.ResponseWriter, r *http.Request) {
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

	var req EditInsightRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErrorWithCode(w, http.StatusBadRequest, "invalid JSON body", "INVALID_JSON", nil)
		return
	}

	result, err := h.editUC.Execute(r.Context(), usecase.EditInsightInput{
		InsightID: insightID,
		CoachID:   claims.Subject,
		Title:     req.Title,
		Body:      req.Body,
	})
	if err != nil {
		handleInsightError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, result)
}

// Share handles PUT /api/v1/insights/{insightID}/share
func (h *InsightHandler) Share(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.shareUC.Execute(r.Context(), usecase.InsightActionInput{
		InsightID: insightID,
		CoachID:   claims.Subject,
	})
	if err != nil {
		handleInsightError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, result)
}

// handleInsightError maps use case errors to HTTP responses.
func handleInsightError(w http.ResponseWriter, err error) {
	if usecase.IsValidationError(err) {
		respondErrorWithCode(w, http.StatusUnprocessableEntity, err.Error(), "VALIDATION_ERROR", nil)
		return
	}
	if usecase.IsAuthenticationError(err) {
		respondErrorWithCode(w, http.StatusForbidden, err.Error(), "FORBIDDEN", nil)
		return
	}
	respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
}
