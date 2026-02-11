package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// PostgresHealthChecker implements the HealthChecker interface for PostgreSQL databases.
// It checks database connectivity by pinging the database with a configurable timeout.
type PostgresHealthChecker struct {
	db      *pgxpool.Pool
	timeout time.Duration
}

// NewPostgresHealthChecker creates a new PostgresHealthChecker instance.
// db is the PostgreSQL connection pool to check.
// timeout is the maximum duration to wait for the database to respond.
func NewPostgresHealthChecker(db *pgxpool.Pool, timeout time.Duration) repository.HealthChecker {
	return &PostgresHealthChecker{
		db:      db,
		timeout: timeout,
	}
}

// Check performs a health check on the PostgreSQL database.
// It pings the database with the configured timeout and returns health status.
// Returns "healthy" status when database is reachable, "degraded" when unreachable.
// This method never returns an error - it uses graceful degradation instead.
func (p *PostgresHealthChecker) Check(ctx context.Context) (*entities.Health, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Measure latency
	start := time.Now()
	err := p.db.Ping(ctx)
	latency := time.Since(start)

	// Build health response based on ping result
	if err != nil {
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
