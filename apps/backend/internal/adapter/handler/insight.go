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

// GetInsightQueueUseCase defines the interface for the insight queue use case.
type GetInsightQueueUseCase interface {
	Execute(ctx context.Context, input usecase.GetInsightQueueInput) (*usecase.GetInsightQueueOutput, error)
}

// GetClientInsightsUseCase defines the interface for the client insights use case.
type GetClientInsightsUseCase interface {
	Execute(ctx context.Context, input usecase.GetClientInsightsInput) (*usecase.GetClientInsightsOutput, error)
}

// InsightActionUseCase is a common interface for single-insight mutation use cases.
type InsightActionUseCase[T any] interface {
	Execute(ctx context.Context, input T) (*entities.InsightCard, error)
}

// InsightHandler handles HTTP requests for insight card operations.
type InsightHandler struct {
	getQueueUC  GetInsightQueueUseCase
	getClientUC GetClientInsightsUseCase
	approveUC   InsightActionUseCase[usecase.ApproveInsightInput]
	dismissUC   InsightActionUseCase[usecase.DismissInsightInput]
	editUC      InsightActionUseCase[usecase.EditInsightInput]
	shareUC     InsightActionUseCase[usecase.ShareInsightInput]
}

// NewInsightHandler creates a new InsightHandler.
func NewInsightHandler(
	getQueueUC GetInsightQueueUseCase,
	getClientUC GetClientInsightsUseCase,
	approveUC InsightActionUseCase[usecase.ApproveInsightInput],
	dismissUC InsightActionUseCase[usecase.DismissInsightInput],
	editUC InsightActionUseCase[usecase.EditInsightInput],
	shareUC InsightActionUseCase[usecase.ShareInsightInput],
) *InsightHandler {
	return &InsightHandler{
		getQueueUC:  getQueueUC,
		getClientUC: getClientUC,
		approveUC:   approveUC,
		dismissUC:   dismissUC,
		editUC:      editUC,
		shareUC:     shareUC,
	}
}

// requireAuth extracts and validates authentication claims from the request.
// Returns nil and writes an error response if authentication fails.
func requireAuth(w http.ResponseWriter, r *http.Request) *middleware.UserClaims {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
	}
	return claims
}

// requireInsightID extracts the insightID URL parameter.
// Returns empty string and writes an error response if missing.
func requireInsightID(w http.ResponseWriter, r *http.Request) string {
	id := chi.URLParam(r, "insightID")
	if id == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing insight ID", "INVALID_REQUEST", nil)
	}
	return id
}

// parsePaginationParams extracts limit, offset, and status from query parameters.
func parsePaginationParams(r *http.Request) (limit, offset int, status string) {
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	status = r.URL.Query().Get("status")
	return
}

// respondInsightList writes a paginated insight list response.
func respondInsightList(w http.ResponseWriter, out *usecase.InsightListOutput) {
	_ = respondJSON(w, http.StatusOK, map[string]interface{}{
		"insights": out.Insights,
		"pagination": map[string]interface{}{
			"page":  (out.Offset / max(out.Limit, 1)) + 1,
			"limit": out.Limit,
			"total": out.Total,
		},
	})
}

// GetQueue handles GET /api/v1/insights/queue
func (h *InsightHandler) GetQueue(w http.ResponseWriter, r *http.Request) {
	claims := requireAuth(w, r)
	if claims == nil {
		return
	}

	if claims.Role != "coach" && claims.Role != "admin" {
		respondErrorWithCode(w, http.StatusForbidden, "only coaches can view the approval queue", "FORBIDDEN", nil)
		return
	}

	limit, offset, status := parsePaginationParams(r)

	out, err := h.getQueueUC.Execute(r.Context(), usecase.GetInsightQueueInput{
		CoachID: claims.Subject,
		Status:  status,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		handleInsightError(w, err)
		return
	}

	respondInsightList(w, out)
}

// GetClientInsights handles GET /api/v1/clients/{clientID}/insights
func (h *InsightHandler) GetClientInsights(w http.ResponseWriter, r *http.Request) {
	claims := requireAuth(w, r)
	if claims == nil {
		return
	}

	clientID := chi.URLParam(r, "clientID")
	if clientID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing client ID", "INVALID_REQUEST", nil)
		return
	}

	limit, offset, status := parsePaginationParams(r)

	out, err := h.getClientUC.Execute(r.Context(), usecase.GetClientInsightsInput{
		CoachID:  claims.Subject,
		ClientID: clientID,
		Status:   status,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		handleInsightError(w, err)
		return
	}

	respondInsightList(w, out)
}

// EditInsightRequest is the JSON body for editing an insight.
type EditInsightRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Approve handles PUT /api/v1/insights/{insightID}/approve
func (h *InsightHandler) Approve(w http.ResponseWriter, r *http.Request) {
	claims := requireAuth(w, r)
	if claims == nil {
		return
	}
	insightID := requireInsightID(w, r)
	if insightID == "" {
		return
	}

	card, err := h.approveUC.Execute(r.Context(), usecase.ApproveInsightInput{
		InsightID: insightID,
		CoachID:   claims.Subject,
	})
	if err != nil {
		handleInsightError(w, err)
		return
	}
	_ = respondJSON(w, http.StatusOK, card)
}

// Dismiss handles PUT /api/v1/insights/{insightID}/dismiss
func (h *InsightHandler) Dismiss(w http.ResponseWriter, r *http.Request) {
	claims := requireAuth(w, r)
	if claims == nil {
		return
	}
	insightID := requireInsightID(w, r)
	if insightID == "" {
		return
	}

	card, err := h.dismissUC.Execute(r.Context(), usecase.DismissInsightInput{
		InsightID: insightID,
		CoachID:   claims.Subject,
	})
	if err != nil {
		handleInsightError(w, err)
		return
	}
	_ = respondJSON(w, http.StatusOK, card)
}

// Edit handles PUT /api/v1/insights/{insightID}
func (h *InsightHandler) Edit(w http.ResponseWriter, r *http.Request) {
	claims := requireAuth(w, r)
	if claims == nil {
		return
	}
	insightID := requireInsightID(w, r)
	if insightID == "" {
		return
	}

	var req EditInsightRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErrorWithCode(w, http.StatusBadRequest, "invalid JSON body", "INVALID_JSON", nil)
		return
	}

	card, err := h.editUC.Execute(r.Context(), usecase.EditInsightInput{
		InsightID: insightID,
		CoachID:   claims.Subject,
		Title:     req.Title,
		Body:      req.Body,
	})
	if err != nil {
		handleInsightError(w, err)
		return
	}
	_ = respondJSON(w, http.StatusOK, card)
}

// Share handles PUT /api/v1/insights/{insightID}/share
func (h *InsightHandler) Share(w http.ResponseWriter, r *http.Request) {
	claims := requireAuth(w, r)
	if claims == nil {
		return
	}
	insightID := requireInsightID(w, r)
	if insightID == "" {
		return
	}

	card, err := h.shareUC.Execute(r.Context(), usecase.ShareInsightInput{
		InsightID: insightID,
		CoachID:   claims.Subject,
	})
	if err != nil {
		handleInsightError(w, err)
		return
	}
	_ = respondJSON(w, http.StatusOK, card)
}

// handleInsightError maps use case errors to appropriate HTTP responses.
func handleInsightError(w http.ResponseWriter, err error) {
	if usecase.IsValidationError(err) {
		respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
		return
	}
	if usecase.IsAuthorizationError(err) {
		respondErrorWithCode(w, http.StatusForbidden, err.Error(), "FORBIDDEN", nil)
		return
	}
	if _, ok := err.(*entities.StatusTransitionError); ok {
		respondErrorWithCode(w, http.StatusUnprocessableEntity, err.Error(), "INVALID_STATUS_TRANSITION", nil)
		return
	}
	respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
}
