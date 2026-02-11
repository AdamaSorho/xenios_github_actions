package entities

import (
	"encoding/json"
	"testing"
)

func TestHealth_Creation(t *testing.T) {
	// Arrange & Act
	health := Health{
		Status: "healthy",
		Checks: map[string]HealthCheck{
			"database": {Status: "up", LatencyMs: 15},
		},
		Uptime: "2h35m",
	}

	// Assert
	if health.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", health.Status)
	}

	if health.Uptime != "2h35m" {
		t.Errorf("expected uptime '2h35m', got '%s'", health.Uptime)
	}

	dbCheck, exists := health.Checks["database"]
	if !exists {
		t.Fatal("expected 'database' check to exist")
	}

	if dbCheck.Status != "up" {
		t.Errorf("expected database status 'up', got '%s'", dbCheck.Status)
	}

	if dbCheck.LatencyMs != 15 {
		t.Errorf("expected latency 15, got %d", dbCheck.LatencyMs)
	}
}

func TestHealth_JSONSerialization(t *testing.T) {
	// Arrange
	health := Health{
		Status: "healthy",
		Checks: map[string]HealthCheck{
			"database": {Status: "up", LatencyMs: 5},
		},
		Uptime: "1h0m",
	}

	// Act
	data, err := json.Marshal(health)
	if err != nil {
		t.Fatalf("failed to marshal health: %v", err)
	}

	var decoded Health
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("failed to unmarshal health: %v", err)
	}

	// Assert
	if decoded.Status != health.Status {
		t.Errorf("expected status '%s', got '%s'", health.Status, decoded.Status)
	}

	if decoded.Uptime != health.Uptime {
		t.Errorf("expected uptime '%s', got '%s'", health.Uptime, decoded.Uptime)
	}

	dbCheck, exists := decoded.Checks["database"]
	if !exists {
		t.Fatal("expected 'database' check to exist after deserialization")
	}

	if dbCheck.LatencyMs != 5 {
		t.Errorf("expected latency 5, got %d", dbCheck.LatencyMs)
	}
}

func TestHealth_JSONFieldNames(t *testing.T) {
	// Arrange
	health := Health{
		Status: "degraded",
		Checks: map[string]HealthCheck{
			"database": {Status: "down", LatencyMs: 0},
		},
		Uptime: "5m",
	}

	// Act
	data, err := json.Marshal(health)
	if err != nil {
		t.Fatalf("failed to marshal health: %v", err)
	}

	jsonStr := string(data)

	// Assert - verify JSON field names match expected format
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMap); err != nil {
		t.Fatalf("failed to unmarshal as raw map: %v", err)
	}

	if _, ok := rawMap["status"]; !ok {
		t.Errorf("expected 'status' field in JSON, got: %s", jsonStr)
	}

	if _, ok := rawMap["checks"]; !ok {
		t.Errorf("expected 'checks' field in JSON, got: %s", jsonStr)
	}

	if _, ok := rawMap["uptime"]; !ok {
		t.Errorf("expected 'uptime' field in JSON, got: %s", jsonStr)
	}
}

func TestHealthCheck_Creation(t *testing.T) {
	// Arrange & Act
	check := HealthCheck{
		Status:    "up",
		LatencyMs: 42,
	}

	// Assert
	if check.Status != "up" {
		t.Errorf("expected status 'up', got '%s'", check.Status)
	}

	if check.LatencyMs != 42 {
		t.Errorf("expected latency 42, got %d", check.LatencyMs)
	}
}

func TestHealthCheck_JSONSerialization(t *testing.T) {
	// Arrange
	check := HealthCheck{
		Status:    "down",
		LatencyMs: 0,
	}

	// Act
	data, err := json.Marshal(check)
	if err != nil {
		t.Fatalf("failed to marshal health check: %v", err)
	}

	var decoded HealthCheck
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("failed to unmarshal health check: %v", err)
	}

	// Assert
	if decoded.Status != check.Status {
		t.Errorf("expected status '%s', got '%s'", check.Status, decoded.Status)
	}

	if decoded.LatencyMs != check.LatencyMs {
		t.Errorf("expected latency %d, got %d", check.LatencyMs, decoded.LatencyMs)
	}
}

func TestHealthCheck_JSONFieldNames(t *testing.T) {
	// Arrange
	check := HealthCheck{
		Status:    "up",
		LatencyMs: 10,
	}

	// Act
	data, err := json.Marshal(check)
	if err != nil {
		t.Fatalf("failed to marshal health check: %v", err)
	}

	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMap); err != nil {
		t.Fatalf("failed to unmarshal as raw map: %v", err)
	}

	// Assert - verify JSON field names
	if _, ok := rawMap["status"]; !ok {
		t.Error("expected 'status' field in JSON")
	}

	if _, ok := rawMap["latency_ms"]; !ok {
		t.Error("expected 'latency_ms' field in JSON")
	}
}

func TestHealth_EmptyChecks(t *testing.T) {
	// Arrange & Act
	health := Health{
		Status: "healthy",
		Checks: map[string]HealthCheck{},
		Uptime: "0s",
	}

	// Assert
	if len(health.Checks) != 0 {
		t.Errorf("expected empty checks map, got %d entries", len(health.Checks))
	}
}

func TestHealth_NilChecks(t *testing.T) {
	// Arrange & Act
	health := Health{
		Status: "healthy",
		Checks: nil,
		Uptime: "0s",
	}

	// Assert
	if health.Checks != nil {
		t.Error("expected nil checks map")
	}
}

func TestHealth_MultipleChecks(t *testing.T) {
	// Arrange & Act
	health := Health{
		Status: "degraded",
		Checks: map[string]HealthCheck{
			"database": {Status: "up", LatencyMs: 5},
			"cache":    {Status: "down", LatencyMs: 0},
			"queue":    {Status: "up", LatencyMs: 12},
		},
		Uptime: "1h30m",
	}

	// Assert
	if len(health.Checks) != 3 {
		t.Errorf("expected 3 checks, got %d", len(health.Checks))
	}

	if health.Checks["database"].Status != "up" {
		t.Errorf("expected database status 'up', got '%s'", health.Checks["database"].Status)
	}

	if health.Checks["cache"].Status != "down" {
		t.Errorf("expected cache status 'down', got '%s'", health.Checks["cache"].Status)
	}
}

func TestHealth_DegradedStatus(t *testing.T) {
	// Arrange & Act
	health := Health{
		Status: "degraded",
		Checks: map[string]HealthCheck{
			"database": {Status: "down", LatencyMs: 0},
		},
		Uptime: "30s",
	}

	// Assert
	if health.Status != "degraded" {
		t.Errorf("expected status 'degraded', got '%s'", health.Status)
	}

	if health.Checks["database"].LatencyMs != 0 {
		t.Errorf("expected latency 0 for down database, got %d", health.Checks["database"].LatencyMs)
	}
}

func TestHealthCheck_ZeroValue(t *testing.T) {
	// Arrange & Act
	var check HealthCheck

	// Assert - zero value should have empty status and 0 latency
	if check.Status != "" {
		t.Errorf("expected empty status for zero value, got '%s'", check.Status)
	}

	if check.LatencyMs != 0 {
		t.Errorf("expected 0 latency for zero value, got %d", check.LatencyMs)
	}
}
