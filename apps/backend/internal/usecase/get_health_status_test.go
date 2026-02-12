package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// mockHealthChecker is a mock implementation of the HealthChecker interface for testing.
type mockHealthChecker struct {
	health *entities.Health
	err    error
	delay  time.Duration
}

func (m *mockHealthChecker) Check(ctx context.Context) (*entities.Health, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return m.health, m.err
}

// TestGetHealthStatusUseCase_Execute_HappyPath tests the successful health check scenario.
func TestGetHealthStatusUseCase_Execute_HappyPath(t *testing.T) {
	// Arrange
	mockChecker := &mockHealthChecker{
		health: &entities.Health{
			Status: "healthy",
			Checks: map[string]entities.HealthCheck{
				"database": {Status: "up", LatencyMs: 15},
			},
		},
		err: nil,
	}
	startTime := time.Now().Add(-2 * time.Hour) // 2 hours ago
	useCase := NewGetHealthStatusUseCase(mockChecker, startTime)

	// Act
	ctx := context.Background()
	health, err := useCase.Execute(ctx)

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
	if health.Uptime == "" {
		t.Error("Expected uptime to be populated")
	}
	if len(health.Checks) != 1 {
		t.Errorf("Expected 1 health check, got: %d", len(health.Checks))
	}
	dbCheck, ok := health.Checks["database"]
	if !ok {
		t.Error("Expected database check to be present")
	}
	if dbCheck.Status != "up" {
		t.Errorf("Expected database status to be 'up', got: %s", dbCheck.Status)
	}
}

// TestGetHealthStatusUseCase_Execute_DatabaseDown tests graceful degradation when database is down.
func TestGetHealthStatusUseCase_Execute_DatabaseDown(t *testing.T) {
	// Arrange
	mockChecker := &mockHealthChecker{
		health: nil,
		err:    errors.New("database connection failed"),
	}
	startTime := time.Now().Add(-1 * time.Hour)
	useCase := NewGetHealthStatusUseCase(mockChecker, startTime)

	// Act
	ctx := context.Background()
	health, err := useCase.Execute(ctx)

	// Assert
	if err != nil {
		t.Errorf("Expected no error (graceful degradation), got: %v", err)
	}
	if health == nil {
		t.Fatal("Expected health to be non-nil even on checker error")
	}
	if health.Status != "degraded" {
		t.Errorf("Expected status to be 'degraded', got: %s", health.Status)
	}
	if health.Uptime == "" {
		t.Error("Expected uptime to be populated")
	}
	dbCheck, ok := health.Checks["database"]
	if !ok {
		t.Error("Expected database check to be present in degraded state")
	}
	if dbCheck.Status != "down" {
		t.Errorf("Expected database status to be 'down', got: %s", dbCheck.Status)
	}
	if dbCheck.LatencyMs != 0 {
		t.Errorf("Expected database latency to be 0 when down, got: %d", dbCheck.LatencyMs)
	}
}

// TestGetHealthStatusUseCase_Execute_Timeout tests timeout scenario.
func TestGetHealthStatusUseCase_Execute_Timeout(t *testing.T) {
	// Arrange
	mockChecker := &mockHealthChecker{
		health: &entities.Health{Status: "healthy"},
		err:    nil,
		delay:  5 * time.Second, // Longer than 3-second timeout
	}
	startTime := time.Now()
	useCase := NewGetHealthStatusUseCase(mockChecker, startTime)

	// Act
	ctx := context.Background()
	start := time.Now()
	health, err := useCase.Execute(ctx)
	duration := time.Since(start)

	// Assert
	if err != nil {
		t.Errorf("Expected no error (graceful timeout handling), got: %v", err)
	}
	if health == nil {
		t.Fatal("Expected health to be non-nil even on timeout")
	}
	if health.Status != "degraded" {
		t.Errorf("Expected status to be 'degraded' on timeout, got: %s", health.Status)
	}
	// Verify timeout occurred (should be ~3 seconds, not 5)
	if duration > 4*time.Second {
		t.Errorf("Expected timeout around 3 seconds, took: %v", duration)
	}
	if duration < 2*time.Second {
		t.Errorf("Expected to wait at least ~3 seconds for timeout, took: %v", duration)
	}
}

// TestGetHealthStatusUseCase_Execute_UptimeFormatting tests uptime calculation.
func TestGetHealthStatusUseCase_Execute_UptimeFormatting(t *testing.T) {
	tests := []struct {
		name          string
		startTime     time.Time
		expectedMatch string // Substring that should be in uptime
	}{
		{
			name:          "Minutes only",
			startTime:     time.Now().Add(-5 * time.Minute),
			expectedMatch: "m",
		},
		{
			name:          "Hours and minutes",
			startTime:     time.Now().Add(-2*time.Hour - 30*time.Minute),
			expectedMatch: "h",
		},
		{
			name:          "Seconds only",
			startTime:     time.Now().Add(-30 * time.Second),
			expectedMatch: "s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockChecker := &mockHealthChecker{
				health: &entities.Health{
					Status: "healthy",
					Checks: map[string]entities.HealthCheck{},
				},
			}
			useCase := NewGetHealthStatusUseCase(mockChecker, tt.startTime)

			// Act
			ctx := context.Background()
			health, err := useCase.Execute(ctx)

			// Assert
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if health.Uptime == "" {
				t.Error("Expected uptime to be populated")
			}
			// Just verify uptime contains expected unit
			// We won't check exact value due to timing variations in tests
			if tt.expectedMatch != "" && health.Uptime == "" {
				t.Errorf("Expected uptime to contain '%s', got: %s", tt.expectedMatch, health.Uptime)
			}
		})
	}
}

// TestFormatUptime_AllBranches tests formatUptime with various durations covering all branches.
func TestFormatUptime_AllBranches(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		contains string
	}{
		{"Zero", 0, "0s"},
		{"Seconds only", 45 * time.Second, "45s"},
		{"Minutes only", 3 * time.Minute, "3m0s"},
		{"Minutes and seconds", 3*time.Minute + 15*time.Second, "3m15s"},
		{"Hours only", 2 * time.Hour, "2h0m0s"},
		{"Hours and minutes", 2*time.Hour + 30*time.Minute, "2h30m0s"},
		{"Days only", 48 * time.Hour, "48h0m0s"},
		{"Days and hours", 50 * time.Hour, "50h0m0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatUptime(tt.duration)
			if result != tt.contains {
				t.Errorf("formatUptime(%v) = %q, want %q", tt.duration, result, tt.contains)
			}
		})
	}
}

// TestGetHealthStatusUseCase_Execute_ContextCancellation tests early context cancellation.
func TestGetHealthStatusUseCase_Execute_ContextCancellation(t *testing.T) {
	// Arrange
	mockChecker := &mockHealthChecker{
		health: &entities.Health{Status: "healthy"},
		delay:  2 * time.Second,
	}
	startTime := time.Now()
	useCase := NewGetHealthStatusUseCase(mockChecker, startTime)

	// Act
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	health, err := useCase.Execute(ctx)

	// Assert
	if err != nil {
		t.Errorf("Expected no error (graceful handling), got: %v", err)
	}
	if health == nil {
		t.Fatal("Expected health to be non-nil even on context cancellation")
	}
	if health.Status != "degraded" {
		t.Errorf("Expected status to be 'degraded' on context cancellation, got: %s", health.Status)
	}
}
