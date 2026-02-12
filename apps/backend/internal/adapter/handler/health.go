package handler

import (
	"context"
	"net/http"

	"github.com/xenios/backend/internal/domain/entities"
)

// GetHealthStatusUseCase defines the interface for the GetHealthStatus use case.
type GetHealthStatusUseCase interface {
	Execute(ctx context.Context) (*entities.Health, error)
}

// HealthHandler handles HTTP requests for health checks.
type HealthHandler struct {
	useCase GetHealthStatusUseCase
}

// NewHealthHandler creates a new HealthHandler with no use case (backward compatibility).
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		useCase: nil,
	}
}

// NewHealthHandlerWithUseCase creates a new HealthHandler with the GetHealthStatus use case.
func NewHealthHandlerWithUseCase(useCase GetHealthStatusUseCase) *HealthHandler {
	return &HealthHandler{
		useCase: useCase,
	}
}

// HealthResponse is the JSON response format for health checks (legacy format).
type HealthResponse struct {
	Status string `json:"status"`
}

// Health handles GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	// If use case is not set, return legacy response for backward compatibility
	if h.useCase == nil {
		_ = respondJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
		return
	}

	// Use the GetHealthStatus use case
	health, err := h.useCase.Execute(r.Context())
	if err != nil {
		// This should never happen as the use case uses graceful degradation,
		// but handle it just in case
		_ = respondJSON(w, http.StatusOK, &entities.Health{
			Status: "degraded",
			Checks: map[string]entities.HealthCheck{
				"system": {Status: "down", LatencyMs: 0},
			},
			Uptime: "unknown",
		})
		return
	}

	// Always return HTTP 200, even when degraded
	_ = respondJSON(w, http.StatusOK, health)
}
