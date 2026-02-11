package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// PostgresHealthChecker implements the HealthChecker interface for PostgreSQL databases.
// It performs health checks by pinging the database connection pool.
type PostgresHealthChecker struct {
	db      *pgxpool.Pool
	timeout time.Duration
}

// NewPostgresHealthChecker creates a new PostgresHealthChecker instance.
// db is the PostgreSQL connection pool to check.
// timeout is the maximum time to wait for the health check to complete.
func NewPostgresHealthChecker(db *pgxpool.Pool, timeout time.Duration) repository.HealthChecker {
	return &PostgresHealthChecker{
		db:      db,
		timeout: timeout,
	}
}

// Check performs a health check on the PostgreSQL database.
// It pings the database with the configured timeout.
// Returns a Health entity with status "healthy" if database is reachable,
// or "degraded" if database is unreachable or times out.
// This method never returns an error - it always returns a Health entity.
func (p *PostgresHealthChecker) Check(ctx context.Context) (*entities.Health, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Measure latency
	start := time.Now()
	err := p.db.Ping(ctx)
	latency := time.Since(start)

	// Build health response
	if err != nil {
		// Database is down or unreachable
		return &entities.Health{
			Status: "degraded",
			Checks: map[string]entities.HealthCheck{
				"database": {
					Status:    "down",
					LatencyMs: 0,
				},
			},
		}, nil
	}

	// Database is healthy
	return &entities.Health{
		Status: "healthy",
		Checks: map[string]entities.HealthCheck{
			"database": {
				Status:    "up",
				LatencyMs: latency.Milliseconds(),
			},
		},
	}, nil
}
