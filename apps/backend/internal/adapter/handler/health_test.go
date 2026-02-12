package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestHealthHandler_Health_Success(t *testing.T) {
	// Arrange
	handler := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Health(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var response HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", response.Status)
	}
}

func TestHealthHandler_Health_ResponseFormat(t *testing.T) {
	// Arrange
	handler := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Health(rec, req)

	// Assert - verify exact JSON structure
	expected := `{"status":"ok"}`
	// Remove trailing newline from encoder
	actual := rec.Body.String()
	if actual != expected+"\n" && actual != expected {
		t.Errorf("expected response body %q, got %q", expected, actual)
	}
}

// TestHealthHandler_Health_WithUseCase_Healthy tests the handler with a mock use case returning healthy status.
func TestHealthHandler_Health_WithUseCase_Healthy(t *testing.T) {
	// Arrange
	mockUseCase := &mockGetHealthStatusUseCase{
		executeFunc: func(ctx context.Context) (*entities.Health, error) {
			return &entities.Health{
				Status: "healthy",
				Checks: map[string]entities.HealthCheck{
					"database": {Status: "up", LatencyMs: 15},
				},
				Uptime: "2h30m",
			}, nil
		},
	}
	handler := NewHealthHandlerWithUseCase(mockUseCase)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Health(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response entities.Health
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", response.Status)
	}

	if response.Checks == nil {
		t.Fatal("expected checks to be non-nil")
	}

	dbCheck, exists := response.Checks["database"]
	if !exists {
		t.Fatal("expected 'database' check to exist")
	}

	if dbCheck.Status != "up" {
		t.Errorf("expected database status 'up', got '%s'", dbCheck.Status)
	}

	if dbCheck.LatencyMs != 15 {
		t.Errorf("expected latency 15ms, got %d", dbCheck.LatencyMs)
	}

	if response.Uptime != "2h30m" {
		t.Errorf("expected uptime '2h30m', got '%s'", response.Uptime)
	}
}

// TestHealthHandler_Health_WithUseCase_Degraded tests the handler with a mock use case returning degraded status.
func TestHealthHandler_Health_WithUseCase_Degraded(t *testing.T) {
	// Arrange
	mockUseCase := &mockGetHealthStatusUseCase{
		executeFunc: func(ctx context.Context) (*entities.Health, error) {
			return &entities.Health{
				Status: "degraded",
				Checks: map[string]entities.HealthCheck{
					"database": {Status: "down", LatencyMs: 0},
				},
				Uptime: "5m",
			}, nil
		},
	}
	handler := NewHealthHandlerWithUseCase(mockUseCase)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Health(rec, req)

	// Assert - always returns HTTP 200, even when degraded
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response entities.Health
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "degraded" {
		t.Errorf("expected status 'degraded', got '%s'", response.Status)
	}

	if response.Checks == nil {
		t.Fatal("expected checks to be non-nil")
	}

	dbCheck, exists := response.Checks["database"]
	if !exists {
		t.Fatal("expected 'database' check to exist")
	}

	if dbCheck.Status != "down" {
		t.Errorf("expected database status 'down', got '%s'", dbCheck.Status)
	}

	if response.Uptime != "5m" {
		t.Errorf("expected uptime '5m', got '%s'", response.Uptime)
	}
}

// TestHealthHandler_Health_WithUseCase_Error tests the handler when the use case returns an error.
func TestHealthHandler_Health_WithUseCase_Error(t *testing.T) {
	mockUseCase := &mockGetHealthStatusUseCase{
		executeFunc: func(ctx context.Context) (*entities.Health, error) {
			return nil, errors.New("unexpected error")
		},
	}
	h := NewHealthHandlerWithUseCase(mockUseCase)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	// Should still return 200 with degraded status
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response entities.Health
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "degraded" {
		t.Errorf("expected status 'degraded', got %q", response.Status)
	}

	systemCheck, exists := response.Checks["system"]
	if !exists {
		t.Fatal("expected 'system' check to exist on error")
	}

	if systemCheck.Status != "down" {
		t.Errorf("expected system status 'down', got %q", systemCheck.Status)
	}

	if response.Uptime != "unknown" {
		t.Errorf("expected uptime 'unknown', got %q", response.Uptime)
	}
}
