package handler

import (
	"context"
	"net/http"

	"github.com/xenios/backend/internal/domain/entities"
)

// GetHealthStatusExecutor defines the interface for executing health status checks.
// This allows for dependency injection and easier testing.
type GetHealthStatusExecutor interface {
	Execute(ctx context.Context) (*entities.Health, error)
}

// HealthHandler handles HTTP requests for health checks.
type HealthHandler struct {
	getHealthStatus GetHealthStatusExecutor
}

// NewHealthHandler creates a new HealthHandler with backward compatibility.
// Returns a handler that returns a simple "ok" status.
// Deprecated: Use NewHealthHandlerWithUseCase for full health checking.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		getHealthStatus: nil,
	}
}

// NewHealthHandlerWithUseCase creates a new HealthHandler with dependency injection.
// getHealthStatus is the use case that performs actual health checks.
func NewHealthHandlerWithUseCase(getHealthStatus GetHealthStatusExecutor) *HealthHandler {
	return &HealthHandler{
		getHealthStatus: getHealthStatus,
	}
}

// HealthResponse is the JSON response format for health checks (legacy).
type HealthResponse struct {
	Status string `json:"status"`
}

// Health handles GET /health
// If a GetHealthStatus use case is configured, it executes health checks.
// Otherwise, it returns a simple "ok" response for backward compatibility.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	// If use case is not configured, return simple response (backward compatibility)
	if h.getHealthStatus == nil {
		respondJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
		return
	}

	// Execute health check use case
	ctx := r.Context()
	health, err := h.getHealthStatus.Execute(ctx)

	// Per use case contract, err should always be nil and health should always be non-nil
	// But handle edge cases gracefully
	if err != nil || health == nil {
		respondJSON(w, http.StatusOK, &entities.Health{
			Status: "degraded",
			Checks: map[string]entities.HealthCheck{},
			Uptime: "",
		})
		return
	}

	// Always return 200 OK, even if status is "degraded"
	// This allows load balancers to distinguish between "service down" and "service degraded"
	respondJSON(w, http.StatusOK, health)
}
