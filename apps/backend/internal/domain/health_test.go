package domain

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

// TestHealthCheck_Creation tests creating a HealthCheck struct
func TestHealthCheck_Creation(t *testing.T) {
	check := HealthCheck{
		Status:    StatusUp,
		LatencyMs: 42,
	}

	if check.Status != StatusUp {
		t.Errorf("Expected Status %s, got %s", StatusUp, check.Status)
	}
	if check.LatencyMs != 42 {
		t.Errorf("Expected LatencyMs 42, got %d", check.LatencyMs)
	}
}

// TestHealth_Creation tests creating a Health entity
func TestHealth_Creation(t *testing.T) {
	checks := map[string]HealthCheck{
		"database": {
			Status:    StatusUp,
			LatencyMs: 15,
		},
		"cache": {
			Status:    StatusDown,
			LatencyMs: 0,
		},
	}

	health := Health{
		Status: HealthStatusDegraded,
		Checks: checks,
		Uptime: "2h35m",
	}

	if health.Status != HealthStatusDegraded {
		t.Errorf("Expected Status %s, got %s", HealthStatusDegraded, health.Status)
	}
	if len(health.Checks) != 2 {
		t.Errorf("Expected 2 checks, got %d", len(health.Checks))
	}
	if health.Uptime != "2h35m" {
		t.Errorf("Expected Uptime '2h35m', got '%s'", health.Uptime)
	}
}

// TestHealth_JSONSerialization tests JSON marshaling of Health entity
func TestHealth_JSONSerialization(t *testing.T) {
	health := Health{
		Status: HealthStatusHealthy,
		Checks: map[string]HealthCheck{
			"database": {
				Status:    StatusUp,
				LatencyMs: 15,
			},
		},
		Uptime: "1h30m",
	}

	jsonBytes, err := json.Marshal(health)
	if err != nil {
		t.Fatalf("Failed to marshal Health to JSON: %v", err)
	}

	var unmarshaled Health
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal Health from JSON: %v", err)
	}

	if unmarshaled.Status != health.Status {
		t.Errorf("Expected Status %s, got %s", health.Status, unmarshaled.Status)
	}
	if unmarshaled.Uptime != health.Uptime {
		t.Errorf("Expected Uptime %s, got %s", health.Uptime, unmarshaled.Uptime)
	}
	if len(unmarshaled.Checks) != len(health.Checks) {
		t.Errorf("Expected %d checks, got %d", len(health.Checks), len(unmarshaled.Checks))
	}

	dbCheck, exists := unmarshaled.Checks["database"]
	if !exists {
		t.Error("Expected 'database' check to exist")
	}
	if dbCheck.Status != StatusUp {
		t.Errorf("Expected database Status %s, got %s", StatusUp, dbCheck.Status)
	}
	if dbCheck.LatencyMs != 15 {
		t.Errorf("Expected database LatencyMs 15, got %d", dbCheck.LatencyMs)
	}
}

// TestHealth_EmptyChecks tests Health entity with empty checks map
func TestHealth_EmptyChecks(t *testing.T) {
	health := Health{
		Status: HealthStatusHealthy,
		Checks: map[string]HealthCheck{},
		Uptime: "5m",
	}

	if health.Status != HealthStatusHealthy {
		t.Errorf("Expected Status %s, got %s", HealthStatusHealthy, health.Status)
	}
	if len(health.Checks) != 0 {
		t.Errorf("Expected 0 checks, got %d", len(health.Checks))
	}
}

// TestHealth_NilChecks tests Health entity with nil checks map
func TestHealth_NilChecks(t *testing.T) {
	health := Health{
		Status: HealthStatusHealthy,
		Checks: nil,
		Uptime: "5m",
	}

	jsonBytes, err := json.Marshal(health)
	if err != nil {
		t.Fatalf("Failed to marshal Health with nil Checks: %v", err)
	}

	// Verify that nil checks marshals to null (not an empty object)
	var raw map[string]interface{}
	err = json.Unmarshal(jsonBytes, &raw)
	if err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	checksValue, exists := raw["checks"]
	if !exists {
		t.Error("Expected 'checks' field in JSON")
	}
	if checksValue != nil {
		t.Errorf("Expected checks to be null, got %v", checksValue)
	}
}

// TestHealthChecker_Interface verifies the interface can be implemented
func TestHealthChecker_Interface(t *testing.T) {
	// This test verifies that the interface is properly defined
	// by attempting to use it in a function signature
	var _ HealthChecker = (*mockHealthChecker)(nil)
}

// mockHealthChecker is a mock implementation for testing
type mockHealthChecker struct {
	shouldError bool
}

func (m *mockHealthChecker) Check(ctx context.Context) (*Health, error) {
	if m.shouldError {
		return nil, context.DeadlineExceeded
	}
	return &Health{
		Status: HealthStatusHealthy,
		Checks: map[string]HealthCheck{
			"mock": {
				Status:    StatusUp,
				LatencyMs: 1,
			},
		},
		Uptime: "0s",
	}, nil
}

// TestHealthChecker_MockImplementation tests using the mock implementation
func TestHealthChecker_MockImplementation(t *testing.T) {
	checker := &mockHealthChecker{}
	ctx := context.Background()

	health, err := checker.Check(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if health == nil {
		t.Fatal("Expected non-nil Health")
	}
	if health.Status != HealthStatusHealthy {
		t.Errorf("Expected Status %s, got %s", HealthStatusHealthy, health.Status)
	}
}

// TestHealthConstants tests that all constants are properly defined
func TestHealthConstants(t *testing.T) {
	// Test Health status constants
	if HealthStatusHealthy != "healthy" {
		t.Errorf("Expected HealthStatusHealthy to be 'healthy', got '%s'", HealthStatusHealthy)
	}
	if HealthStatusDegraded != "degraded" {
		t.Errorf("Expected HealthStatusDegraded to be 'degraded', got '%s'", HealthStatusDegraded)
	}

	// Test HealthCheck status constants
	if StatusUp != "up" {
		t.Errorf("Expected StatusUp to be 'up', got '%s'", StatusUp)
	}
	if StatusDown != "down" {
		t.Errorf("Expected StatusDown to be 'down', got '%s'", StatusDown)
	}
}

// TestHealth_MultipleChecks tests Health with various check combinations
func TestHealth_MultipleChecks(t *testing.T) {
	tests := []struct {
		name           string
		health         Health
		expectedStatus string
		expectedCount  int
	}{
		{
			name: "all checks up",
			health: Health{
				Status: HealthStatusHealthy,
				Checks: map[string]HealthCheck{
					"database": {Status: StatusUp, LatencyMs: 10},
					"cache":    {Status: StatusUp, LatencyMs: 5},
					"storage":  {Status: StatusUp, LatencyMs: 20},
				},
				Uptime: "1h",
			},
			expectedStatus: HealthStatusHealthy,
			expectedCount:  3,
		},
		{
			name: "mixed checks",
			health: Health{
				Status: HealthStatusDegraded,
				Checks: map[string]HealthCheck{
					"database": {Status: StatusUp, LatencyMs: 10},
					"cache":    {Status: StatusDown, LatencyMs: 0},
				},
				Uptime: "30m",
			},
			expectedStatus: HealthStatusDegraded,
			expectedCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.health.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, tt.health.Status)
			}
			if len(tt.health.Checks) != tt.expectedCount {
				t.Errorf("Expected %d checks, got %d", tt.expectedCount, len(tt.health.Checks))
			}
		})
	}
}

// TestHealthChecker_CheckError tests that HealthChecker.Check() can return errors
func TestHealthChecker_CheckError(t *testing.T) {
	checker := &mockHealthChecker{shouldError: true}
	ctx := context.Background()

	health, err := checker.Check(ctx)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if health != nil {
		t.Errorf("Expected nil Health when error occurs, got %v", health)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded error, got %v", err)
	}
}

// TestHealthChecker_ContextCancellation tests context timeout handling
func TestHealthChecker_ContextCancellation(t *testing.T) {
	checker := &mockHealthChecker{}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(5 * time.Millisecond)

	// Verify context is expired
	if ctx.Err() == nil {
		t.Fatal("Expected context to be expired")
	}
	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got %v", ctx.Err())
	}

	// The mock doesn't actually respect context, but we verify the pattern works
	// In real implementations, Check() should respect ctx.Done()
	_, err := checker.Check(ctx)
	// Mock returns nil error, but in production code should check ctx.Err()
	if err != nil {
		t.Logf("Mock returned error (this is OK): %v", err)
	}
}

// TestHealthCheck_NegativeLatency tests HealthCheck with negative latency values
func TestHealthCheck_NegativeLatency(t *testing.T) {
	check := HealthCheck{
		Status:    StatusDown,
		LatencyMs: -1,
	}

	if check.LatencyMs >= 0 {
		t.Errorf("Expected negative LatencyMs, got %d", check.LatencyMs)
	}

	// Verify negative latency can be JSON marshaled/unmarshaled
	jsonBytes, err := json.Marshal(check)
	if err != nil {
		t.Fatalf("Failed to marshal HealthCheck with negative latency: %v", err)
	}

	var unmarshaled HealthCheck
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal HealthCheck with negative latency: %v", err)
	}

	if unmarshaled.LatencyMs != -1 {
		t.Errorf("Expected LatencyMs -1, got %d", unmarshaled.LatencyMs)
	}
}

// TestHealth_EmptyStrings tests Health entity with empty string values
func TestHealth_EmptyStrings(t *testing.T) {
	health := Health{
		Status: "",
		Checks: map[string]HealthCheck{
			"database": {
				Status:    "",
				LatencyMs: 10,
			},
		},
		Uptime: "",
	}

	if health.Status != "" {
		t.Errorf("Expected empty Status, got %s", health.Status)
	}
	if health.Uptime != "" {
		t.Errorf("Expected empty Uptime, got %s", health.Uptime)
	}

	dbCheck, exists := health.Checks["database"]
	if !exists {
		t.Fatal("Expected database check to exist")
	}
	if dbCheck.Status != "" {
		t.Errorf("Expected empty check Status, got %s", dbCheck.Status)
	}

	// Verify empty strings serialize correctly
	jsonBytes, err := json.Marshal(health)
	if err != nil {
		t.Fatalf("Failed to marshal Health with empty strings: %v", err)
	}

	var unmarshaled Health
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal Health with empty strings: %v", err)
	}

	if unmarshaled.Status != "" {
		t.Errorf("Expected empty Status after unmarshal, got %s", unmarshaled.Status)
	}
	if unmarshaled.Uptime != "" {
		t.Errorf("Expected empty Uptime after unmarshal, got %s", unmarshaled.Uptime)
	}
}
