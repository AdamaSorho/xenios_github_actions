package entities

// Health represents the overall health status of the system.
type Health struct {
	Status string                  `json:"status"` // "healthy", "degraded", "down"
	Checks map[string]HealthCheck  `json:"checks"`
	Uptime string                  `json:"uptime"` // Human-readable uptime (e.g., "2h35m")
}

// HealthCheck represents the health status of a single dependency.
type HealthCheck struct {
	Status    string `json:"status"`     // "up" or "down"
	LatencyMs int64  `json:"latency_ms"` // Response time in milliseconds
}
