package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestPostgresHealthChecker_Check_Healthy tests the health checker with a live database connection.
// This is an integration test that requires a real PostgreSQL database.
func TestPostgresHealthChecker_Check_Healthy(t *testing.T) {
	// Skip if no database URL is provided
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	// Arrange
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}
	defer pool.Close()

	checker := NewPostgresHealthChecker(pool, 3*time.Second)

	// Act
	health, err := checker.Check(ctx)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if health == nil {
		t.Fatal("expected health to be non-nil")
	}

	if health.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", health.Status)
	}

	if health.Checks == nil {
		t.Fatal("expected checks to be non-nil")
	}

	dbCheck, exists := health.Checks["database"]
	if !exists {
		t.Fatal("expected 'database' check to exist")
	}

	if dbCheck.Status != "up" {
		t.Errorf("expected database status 'up', got '%s'", dbCheck.Status)
	}

	if dbCheck.LatencyMs <= 0 {
		t.Errorf("expected positive latency, got %d", dbCheck.LatencyMs)
	}
}

// TestPostgresHealthChecker_Check_Timeout tests the health checker with a timeout.
func TestPostgresHealthChecker_Check_Timeout(t *testing.T) {
	// Skip if no database URL is provided
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	// Arrange
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}
	defer pool.Close()

	// Use a very short timeout (1 nanosecond) to force a timeout
	checker := NewPostgresHealthChecker(pool, 1*time.Nanosecond)

	// Act
	health, err := checker.Check(ctx)

	// Assert
	if err != nil {
		t.Fatalf("expected no error (graceful degradation), got: %v", err)
	}

	if health == nil {
		t.Fatal("expected health to be non-nil")
	}

	if health.Status != "degraded" {
		t.Errorf("expected status 'degraded', got '%s'", health.Status)
	}

	dbCheck, exists := health.Checks["database"]
	if !exists {
		t.Fatal("expected 'database' check to exist")
	}

	if dbCheck.Status != "down" {
		t.Errorf("expected database status 'down', got '%s'", dbCheck.Status)
	}
}

// TestPostgresHealthChecker_Check_UnreachableDatabase tests the health checker with a bad connection string.
func TestPostgresHealthChecker_Check_UnreachableDatabase(t *testing.T) {
	// Arrange - use an invalid connection string
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://invalid:invalid@localhost:9999/invalid")
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}
	defer pool.Close()

	checker := NewPostgresHealthChecker(pool, 3*time.Second)

	// Act
	health, err := checker.Check(ctx)

	// Assert
	if err != nil {
		t.Fatalf("expected no error (graceful degradation), got: %v", err)
	}

	if health == nil {
		t.Fatal("expected health to be non-nil")
	}

	if health.Status != "degraded" {
		t.Errorf("expected status 'degraded', got '%s'", health.Status)
	}

	dbCheck, exists := health.Checks["database"]
	if !exists {
		t.Fatal("expected 'database' check to exist")
	}

	if dbCheck.Status != "down" {
		t.Errorf("expected database status 'down', got '%s'", dbCheck.Status)
	}
}
