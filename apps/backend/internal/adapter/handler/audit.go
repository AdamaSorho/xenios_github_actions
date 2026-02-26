package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/xenios/backend/internal/usecase"
)

// QueryAuditLogUseCase defines the interface for the query audit log use case.
type QueryAuditLogUseCase interface {
	Execute(ctx context.Context, input usecase.QueryAuditLogInput) (*usecase.QueryAuditLogOutput, error)
}

// AuditHandler handles HTTP requests for audit log queries.
type AuditHandler struct {
	queryUC QueryAuditLogUseCase
}

// NewAuditHandler creates a new AuditHandler.
func NewAuditHandler(queryUC QueryAuditLogUseCase) *AuditHandler {
	return &AuditHandler{queryUC: queryUC}
}

// Query handles GET /api/v1/admin/audit
func (h *AuditHandler) Query(w http.ResponseWriter, r *http.Request) {
	claims := requireAuth(w, r)
	if claims == nil {
		return
	}

	if claims.Role != "admin" {
		respondErrorWithCode(w, http.StatusForbidden, "admin access required", "FORBIDDEN", nil)
		return
	}

	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	input := usecase.QueryAuditLogInput{
		ActorID:    q.Get("actor_id"),
		Action:     q.Get("action"),
		EntityType: q.Get("entity_type"),
		EntityID:   q.Get("entity_id"),
		Limit:      limit,
		Offset:     offset,
	}

	if fromStr := q.Get("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			input.From = &t
		}
	}
	if toStr := q.Get("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			input.To = &t
		}
	}

	out, err := h.queryUC.Execute(r.Context(), input)
	if err != nil {
		respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
		return
	}

	_ = respondJSON(w, http.StatusOK, out)
}
