package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
