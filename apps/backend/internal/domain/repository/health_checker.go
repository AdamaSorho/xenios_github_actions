package repository

import (
	"context"
	"github.com/xenios/backend/internal/domain/entities"
)

// HealthChecker defines the interface for checking system health.
// This is a repository-like interface that abstracts the health checking logic.
type HealthChecker interface {
	// Check performs health checks on system dependencies.
	// It returns a Health entity with the status of all checks.
	// If the context is cancelled or times out, it should return an error.
	Check(ctx context.Context) (*entities.Health, error)
}
