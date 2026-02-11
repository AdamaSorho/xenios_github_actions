package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

// mockHealthChecker is a test implementation of the HealthChecker interface.
type mockHealthChecker struct {
	health *entities.Health
	err    error
}

func (m *mockHealthChecker) Check(ctx context.Context) (*entities.Health, error) {
	return m.health, m.err
}

func TestHealthChecker_InterfaceCompliance(t *testing.T) {
	// Assert that mockHealthChecker implements HealthChecker
	var _ HealthChecker = (*mockHealthChecker)(nil)
}

func TestHealthChecker_MockReturnsHealthy(t *testing.T) {
	// Arrange
	mock := &mockHealthChecker{
		health: &entities.Health{
			Status: "healthy",
			Checks: map[string]entities.HealthCheck{
				"database": {Status: "up", LatencyMs: 5},
			},
			Uptime: "1h0m",
		},
		err: nil,
	}

	// Act
	health, err := mock.Check(context.Background())

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if health == nil {
		t.Fatal("expected non-nil health")
	}

	if health.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", health.Status)
	}
}

func TestHealthChecker_MockReturnsDegraded(t *testing.T) {
	// Arrange
	mock := &mockHealthChecker{
		health: &entities.Health{
			Status: "degraded",
			Checks: map[string]entities.HealthCheck{
				"database": {Status: "down", LatencyMs: 0},
			},
		},
		err: nil,
	}

	// Act
	health, err := mock.Check(context.Background())

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if health.Status != "degraded" {
		t.Errorf("expected status 'degraded', got '%s'", health.Status)
	}
}

func TestHealthChecker_MockReturnsError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("connection refused")
	mock := &mockHealthChecker{
		health: nil,
		err:    expectedErr,
	}

	// Act
	health, err := mock.Check(context.Background())

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != expectedErr {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}

	if health != nil {
		t.Error("expected nil health when error returned")
	}
}

func TestHealthChecker_MockContextCancellation(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mock := &mockHealthChecker{
		health: &entities.Health{
			Status: "healthy",
			Checks: map[string]entities.HealthCheck{
				"database": {Status: "up", LatencyMs: 5},
			},
		},
		err: nil,
	}

	// Act - the mock doesn't check context, but this validates the interface accepts it
	health, err := mock.Check(ctx)

	// Assert
	if err != nil {
		t.Fatalf("mock doesn't check context, expected no error, got: %v", err)
	}

	if health == nil {
		t.Fatal("expected non-nil health")
	}
}
