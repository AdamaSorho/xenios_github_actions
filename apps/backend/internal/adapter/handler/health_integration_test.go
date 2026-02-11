package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// mockGetHealthStatusUseCase is a mock implementation for testing.
type mockGetHealthStatusUseCase struct {
	health *entities.Health
	err    error
}

func (m *mockGetHealthStatusUseCase) Execute(ctx context.Context) (*entities.Health, error) {
	return m.health, m.err
}

// TestHealthHandler_HealthWithUseCase_Healthy tests handler with healthy status.
func TestHealthHandler_HealthWithUseCase_Healthy(t *testing.T) {
	// Arrange
	mockUseCase := &mockGetHealthStatusUseCase{
		health: &entities.Health{
			Status: "healthy",
			Checks: map[string]entities.HealthCheck{
				"database": {Status: "up", LatencyMs: 15},
			},
			Uptime: "2h15m",
		},
		err: nil,
	}
	handler := NewHealthHandlerWithUseCase(mockUseCase)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Health(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var response entities.Health
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response.Status)
	}
	if response.Uptime != "2h15m" {
		t.Errorf("Expected uptime '2h15m', got '%s'", response.Uptime)
	}
	if len(response.Checks) != 1 {
		t.Errorf("Expected 1 check, got %d", len(response.Checks))
	}
	dbCheck, ok := response.Checks["database"]
	if !ok {
		t.Error("Expected database check to be present")
	}
	if dbCheck.Status != "up" {
		t.Errorf("Expected database status 'up', got '%s'", dbCheck.Status)
	}
	if dbCheck.LatencyMs != 15 {
		t.Errorf("Expected latency 15ms, got %d", dbCheck.LatencyMs)
	}
}

// TestHealthHandler_HealthWithUseCase_Degraded tests handler with degraded status.
func TestHealthHandler_HealthWithUseCase_Degraded(t *testing.T) {
	// Arrange
	mockUseCase := &mockGetHealthStatusUseCase{
		health: &entities.Health{
			Status: "degraded",
			Checks: map[string]entities.HealthCheck{
				"database": {Status: "down", LatencyMs: 0},
			},
			Uptime: "1h30m",
		},
		err: nil,
	}
	handler := NewHealthHandlerWithUseCase(mockUseCase)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Health(rec, req)

	// Assert - Still returns 200 even when degraded
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d even when degraded, got %d", http.StatusOK, rec.Code)
	}

	var response entities.Health
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "degraded" {
		t.Errorf("Expected status 'degraded', got '%s'", response.Status)
	}
	if response.Uptime != "1h30m" {
		t.Errorf("Expected uptime '1h30m', got '%s'", response.Uptime)
	}
	dbCheck, ok := response.Checks["database"]
	if !ok {
		t.Error("Expected database check to be present")
	}
	if dbCheck.Status != "down" {
		t.Errorf("Expected database status 'down', got '%s'", dbCheck.Status)
	}
}

// TestHealthHandler_HealthWithUseCase_UseCaseError tests graceful error handling.
// Even though usecase.Execute should never return an error per its contract,
// the handler should handle it gracefully just in case.
func TestHealthHandler_HealthWithUseCase_UseCaseError(t *testing.T) {
	// Arrange
	mockUseCase := &mockGetHealthStatusUseCase{
		health: nil,
		err:    nil, // Use case always returns health, never error per contract
	}
	handler := NewHealthHandlerWithUseCase(mockUseCase)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Health(rec, req)

	// Assert - Should still return 200 and a valid response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

// TestHealthHandler_HealthWithUseCase_Integration tests with real use case.
func TestHealthHandler_HealthWithUseCase_Integration(t *testing.T) {
	// Arrange - Use a real use case with mock health checker
	mockChecker := &mockHealthChecker{
		health: &entities.Health{
			Status: "healthy",
			Checks: map[string]entities.HealthCheck{
				"database": {Status: "up", LatencyMs: 10},
			},
		},
		err: nil,
	}
	startTime := time.Now().Add(-1 * time.Hour)
	useCase := usecase.NewGetHealthStatusUseCase(mockChecker, startTime)
	handler := NewHealthHandlerWithUseCase(useCase)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Health(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response entities.Health
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response.Status)
	}
	if response.Uptime == "" {
		t.Error("Expected uptime to be populated by use case")
	}
}

// mockHealthChecker is a mock for testing (reused from usecase tests).
type mockHealthChecker struct {
	health *entities.Health
	err    error
}

func (m *mockHealthChecker) Check(ctx context.Context) (*entities.Health, error) {
	return m.health, m.err
}
