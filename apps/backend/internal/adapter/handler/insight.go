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

// Use case interfaces for dependency injection.

// GetInsightQueueUseCaseInterface defines the interface for the get insight queue use case.
type GetInsightQueueUseCaseInterface interface {
	Execute(ctx context.Context, input usecase.GetInsightQueueInput) (*usecase.GetInsightQueueOutput, error)
}

// ApproveInsightUseCaseInterface defines the interface for the approve insight use case.
type ApproveInsightUseCaseInterface interface {
	Execute(ctx context.Context, insightID, coachID string) (*entities.InsightCard, error)
}

// DismissInsightUseCaseInterface defines the interface for the dismiss insight use case.
type DismissInsightUseCaseInterface interface {
	Execute(ctx context.Context, insightID, coachID string) (*entities.InsightCard, error)
}

// EditInsightUseCaseInterface defines the interface for the edit insight use case.
type EditInsightUseCaseInterface interface {
	Execute(ctx context.Context, input usecase.EditInsightInput) (*entities.InsightCard, error)
}

// ShareInsightUseCaseInterface defines the interface for the share insight use case.
type ShareInsightUseCaseInterface interface {
	Execute(ctx context.Context, insightID, coachID string) (*entities.InsightCard, error)
}

// GetClientInsightsUseCaseInterface defines the interface for the get client insights use case.
type GetClientInsightsUseCaseInterface interface {
	Execute(ctx context.Context, input usecase.GetClientInsightsInput) (*usecase.GetClientInsightsOutput, error)
}

// InsightHandler handles HTTP requests for insight card operations.
type InsightHandler struct {
	getQueueUC    GetInsightQueueUseCaseInterface
	approveUC     ApproveInsightUseCaseInterface
	dismissUC     DismissInsightUseCaseInterface
	editUC        EditInsightUseCaseInterface
	shareUC       ShareInsightUseCaseInterface
	getClientUC   GetClientInsightsUseCaseInterface
}

// NewInsightHandler creates a new InsightHandler.
func NewInsightHandler(
	getQueueUC GetInsightQueueUseCaseInterface,
	approveUC ApproveInsightUseCaseInterface,
	dismissUC DismissInsightUseCaseInterface,
	editUC EditInsightUseCaseInterface,
	shareUC ShareInsightUseCaseInterface,
	getClientUC GetClientInsightsUseCaseInterface,
) *InsightHandler {
	return &InsightHandler{
		getQueueUC:  getQueueUC,
		approveUC:   approveUC,
		dismissUC:   dismissUC,
		editUC:      editUC,
		shareUC:     shareUC,
		getClientUC: getClientUC,
	}
}

// GetQueue handles GET /api/v1/insights/queue
func (h *InsightHandler) GetQueue(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	page, _ := strconv.Atoi(q.Get("page"))
	if page > 0 && limit > 0 && offset == 0 {
		offset = (page - 1) * limit
	}

	input := usecase.GetInsightQueueInput{
		CoachID: claims.Subject,
		Status:  q.Get("status"),
		Limit:   limit,
		Offset:  offset,
	}

	out, err := h.getQueueUC.Execute(r.Context(), input)
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

	out, err := h.approveUC.Execute(r.Context(), insightID, claims.Subject)
	if err != nil {
		handleInsightError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
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

	out, err := h.dismissUC.Execute(r.Context(), insightID, claims.Subject)
	if err != nil {
		handleInsightError(w, err)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
}

// EditInsightRequest is the JSON request body for editing an insight card.
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

	out, err := h.shareUC.Execute(r.Context(), insightID, claims.Subject)
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

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	input := usecase.GetClientInsightsInput{
		ClientID: clientID,
		CoachID:  claims.Subject,
		Status:   q.Get("status"),
		Limit:    limit,
		Offset:   offset,
	}

	out, err := h.getClientUC.Execute(r.Context(), input)
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
	if usecase.IsAuthorizationError(err) {
		respondErrorWithCode(w, http.StatusForbidden, err.Error(), "FORBIDDEN", nil)
		return
	}
	if usecase.IsInvalidTransitionError(err) {
		respondErrorWithCode(w, http.StatusUnprocessableEntity, err.Error(), "INVALID_TRANSITION", nil)
		return
	}
	respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
}
