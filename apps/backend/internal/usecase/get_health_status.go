package usecase

import (
	"context"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetHealthStatusUseCase orchestrates health checking across system dependencies.
type GetHealthStatusUseCase struct {
	healthChecker repository.HealthChecker
	startTime     time.Time
}

// NewGetHealthStatusUseCase creates a new GetHealthStatusUseCase instance.
// healthChecker is the dependency that performs actual health checks.
// startTime is the application start time, used to calculate uptime.
func NewGetHealthStatusUseCase(healthChecker repository.HealthChecker, startTime time.Time) *GetHealthStatusUseCase {
	return &GetHealthStatusUseCase{
		healthChecker: healthChecker,
		startTime:     startTime,
	}
}

// Execute performs the health check with a 3-second timeout.
// It returns a Health entity with system status and never propagates errors.
// If the health checker fails or times out, it returns a degraded status.
func (uc *GetHealthStatusUseCase) Execute(ctx context.Context) (*entities.Health, error) {
	// Create context with 3-second timeout
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Call health checker
	health, err := uc.healthChecker.Check(ctx)
	if err != nil {
		// Graceful degradation: return degraded status, not error
		health = &entities.Health{
			Status: "degraded",
			Checks: map[string]entities.HealthCheck{
				"database": {Status: "down", LatencyMs: 0},
			},
			Uptime: formatUptime(time.Since(uc.startTime)),
		}
	} else {
		// Add uptime to the health response
		health.Uptime = formatUptime(time.Since(uc.startTime))
	}

	return health, nil
}

// formatUptime converts a duration to a human-readable uptime string.
// Examples: "5m32s", "2h15m", "3d4h"
func formatUptime(d time.Duration) string {
	d = d.Round(time.Second)

	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute
	d -= minutes * time.Minute
	seconds := d / time.Second

	if days > 0 {
		if hours > 0 {
			return time.Duration(days*24*time.Hour + hours*time.Hour).String()
		}
		return time.Duration(days * 24 * time.Hour).String()
	}
	if hours > 0 {
		if minutes > 0 {
			return time.Duration(hours*time.Hour + minutes*time.Minute).String()
		}
		return time.Duration(hours * time.Hour).String()
	}
	if minutes > 0 {
		if seconds > 0 {
			return time.Duration(minutes*time.Minute + seconds*time.Second).String()
		}
		return time.Duration(minutes * time.Minute).String()
	}
	return time.Duration(seconds * time.Second).String()
}
