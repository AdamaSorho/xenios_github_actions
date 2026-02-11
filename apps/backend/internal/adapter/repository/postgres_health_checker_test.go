package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestPostgresHealthChecker_Check_DatabaseHealthy tests successful database ping.
func TestPostgresHealthChecker_Check_DatabaseHealthy(t *testing.T) {
	// Skip if no database connection available
	// This is an integration test that requires a real database
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Arrange
	ctx := context.Background()
	pool, err := createTestPool(ctx, t)
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
	}
	defer pool.Close()

	checker := NewPostgresHealthChecker(pool, 3*time.Second)

	// Act
	health, err := checker.Check(ctx)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if health == nil {
		t.Fatal("Expected health to be non-nil")
	}
	if health.Status != "healthy" {
		t.Errorf("Expected status to be 'healthy', got: %s", health.Status)
	}
	if health.Checks == nil {
		t.Fatal("Expected checks to be non-nil")
	}
	dbCheck, ok := health.Checks["database"]
	if !ok {
		t.Error("Expected database check to be present")
	}
	if dbCheck.Status != "up" {
		t.Errorf("Expected database status to be 'up', got: %s", dbCheck.Status)
	}
	if dbCheck.LatencyMs < 0 {
		t.Errorf("Expected latency to be non-negative, got: %d", dbCheck.LatencyMs)
	}
}

// TestPostgresHealthChecker_Check_DatabaseUnreachable tests degraded status when database is down.
func TestPostgresHealthChecker_Check_DatabaseUnreachable(t *testing.T) {
	// Arrange - Create pool with invalid connection string
	ctx := context.Background()
	config, err := pgxpool.ParseConfig("postgres://invalid:invalid@localhost:9999/nonexistent")
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}
	config.MaxConns = 1
	config.MinConns = 0
	// Set very short timeout to avoid waiting
	config.ConnConfig.ConnectTimeout = 100 * time.Millisecond

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	checker := NewPostgresHealthChecker(pool, 1*time.Second)

	// Act
	health, err := checker.Check(ctx)

	// Assert
	if err != nil {
		t.Errorf("Expected no error (graceful degradation), got: %v", err)
	}
	if health == nil {
		t.Fatal("Expected health to be non-nil even when database is down")
	}
	if health.Status != "degraded" {
		t.Errorf("Expected status to be 'degraded', got: %s", health.Status)
	}
	if health.Checks == nil {
		t.Fatal("Expected checks to be non-nil")
	}
	dbCheck, ok := health.Checks["database"]
	if !ok {
		t.Error("Expected database check to be present")
	}
	if dbCheck.Status != "down" {
		t.Errorf("Expected database status to be 'down', got: %s", dbCheck.Status)
	}
	if dbCheck.LatencyMs != 0 {
		t.Errorf("Expected latency to be 0 when down, got: %d", dbCheck.LatencyMs)
	}
}

// TestPostgresHealthChecker_Check_Timeout tests timeout scenario.
func TestPostgresHealthChecker_Check_Timeout(t *testing.T) {
	// Skip if no database connection available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Arrange
	ctx := context.Background()
	pool, err := createTestPool(ctx, t)
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
	}
	defer pool.Close()

	// Use very short timeout
	checker := NewPostgresHealthChecker(pool, 1*time.Nanosecond)

	// Act
	start := time.Now()
	health, err := checker.Check(ctx)
	duration := time.Since(start)

	// Assert
	if err != nil {
		t.Errorf("Expected no error (graceful timeout), got: %v", err)
	}
	if health == nil {
		t.Fatal("Expected health to be non-nil even on timeout")
	}
	if health.Status != "degraded" {
		t.Errorf("Expected status to be 'degraded', got: %s", health.Status)
	}
	// Should timeout very quickly
	if duration > 100*time.Millisecond {
		t.Errorf("Expected quick timeout, took: %v", duration)
	}
}

// TestPostgresHealthChecker_Check_ContextCancellation tests context cancellation.
func TestPostgresHealthChecker_Check_ContextCancellation(t *testing.T) {
	// Skip if no database connection available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Arrange
	ctx := context.Background()
	pool, err := createTestPool(ctx, t)
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
	}
	defer pool.Close()

	checker := NewPostgresHealthChecker(pool, 3*time.Second)

	// Act
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately
	health, err := checker.Check(cancelCtx)

	// Assert
	if err != nil {
		t.Errorf("Expected no error (graceful cancellation), got: %v", err)
	}
	if health == nil {
		t.Fatal("Expected health to be non-nil even on cancellation")
	}
	if health.Status != "degraded" {
		t.Errorf("Expected status to be 'degraded', got: %s", health.Status)
	}
}

// createTestPool creates a connection pool for testing.
// It tries to connect to a local PostgreSQL instance.
// If DATABASE_URL is set, it uses that; otherwise it uses default localhost settings.
func createTestPool(ctx context.Context, t *testing.T) (*pgxpool.Pool, error) {
	t.Helper()

	// Try DATABASE_URL from environment first
	connString := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	// Allow override via environment variable for CI/CD
	// if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
	// 	connString = dbURL
	// }

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 2
	config.MinConns = 1
	config.ConnConfig.ConnectTimeout = 2 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
