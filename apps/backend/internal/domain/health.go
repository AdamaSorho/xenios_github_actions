package domain

import "context"

// Health status constants
const (
	// HealthStatusHealthy indicates all dependencies are operational
	HealthStatusHealthy = "healthy"
	// HealthStatusDegraded indicates some dependencies are down but the service is still operational
	HealthStatusDegraded = "degraded"
)

// HealthCheck status constants
const (
	// StatusUp indicates the dependency is operational
	StatusUp = "up"
	// StatusDown indicates the dependency is not operational
	StatusDown = "down"
)

// HealthChecker is the domain interface for health checking.
// Implementations in the infrastructure layer will check actual service dependencies.
type HealthChecker interface {
	// Check performs health checks on service dependencies and returns the overall health status.
	// Returns Health entity with status and individual dependency checks, or an error if the check fails.
	Check(ctx context.Context) (*Health, error)
}

// Health represents the overall health status of the application and its dependencies.
type Health struct {
	// Status is the overall health status: "healthy" or "degraded"
	Status string `json:"status"`
	// Checks is a map of dependency names to their individual health check results
	Checks map[string]HealthCheck `json:"checks"`
	// Uptime is the duration the service has been running (e.g., "2h35m")
	Uptime string `json:"uptime"`
}

// HealthCheck represents the health status of an individual dependency.
type HealthCheck struct {
	// Status is the dependency status: "up" or "down"
	Status string `json:"status"`
	// LatencyMs is the response time in milliseconds
	LatencyMs int64 `json:"latency_ms"`
}
