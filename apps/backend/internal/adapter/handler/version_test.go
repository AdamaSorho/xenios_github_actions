package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"
)

func TestVersionHandler_Version_Success(t *testing.T) {
	// Arrange
	handler := NewVersionHandler()
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Version(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var response VersionResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Version == "" {
		t.Error("expected version to be non-empty")
	}

	if response.Environment == "" {
		t.Error("expected environment to be non-empty")
	}

	if response.GoVersion == "" {
		t.Error("expected goVersion to be non-empty")
	}
}

func TestVersionHandler_Version_ReturnsCorrectVersion(t *testing.T) {
	// Arrange
	handler := NewVersionHandler()
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Version(rec, req)

	// Assert
	var response VersionResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expected := "0.1.0"
	if response.Version != expected {
		t.Errorf("expected version %q, got %q", expected, response.Version)
	}
}

func TestVersionHandler_Version_ReturnsGoVersion(t *testing.T) {
	// Arrange
	handler := NewVersionHandler()
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Version(rec, req)

	// Assert
	var response VersionResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expected := runtime.Version()
	if response.GoVersion != expected {
		t.Errorf("expected goVersion %q, got %q", expected, response.GoVersion)
	}
}

func TestVersionHandler_Version_ReturnsEnvironmentFromEnvVar(t *testing.T) {
	// Arrange
	originalEnv := os.Getenv("ENVIRONMENT")
	os.Setenv("ENVIRONMENT", "staging")
	defer os.Setenv("ENVIRONMENT", originalEnv)

	handler := NewVersionHandler()
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Version(rec, req)

	// Assert
	var response VersionResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expected := "staging"
	if response.Environment != expected {
		t.Errorf("expected environment %q, got %q", expected, response.Environment)
	}
}

func TestVersionHandler_Version_DefaultsToDevEnvironment(t *testing.T) {
	// Arrange
	originalEnv := os.Getenv("ENVIRONMENT")
	os.Unsetenv("ENVIRONMENT")
	defer os.Setenv("ENVIRONMENT", originalEnv)

	handler := NewVersionHandler()
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Version(rec, req)

	// Assert
	var response VersionResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expected := "development"
	if response.Environment != expected {
		t.Errorf("expected environment %q, got %q", expected, response.Environment)
	}
}

func TestVersionHandler_Version_ResponseFormat(t *testing.T) {
	// Arrange
	originalEnv := os.Getenv("ENVIRONMENT")
	os.Setenv("ENVIRONMENT", "test")
	defer os.Setenv("ENVIRONMENT", originalEnv)

	handler := NewVersionHandler()
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Version(rec, req)

	// Assert - verify JSON has all expected fields
	var rawResponse map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&rawResponse); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	requiredFields := []string{"version", "environment", "goVersion"}
	for _, field := range requiredFields {
		if _, exists := rawResponse[field]; !exists {
			t.Errorf("response missing required field %q", field)
		}
	}

	// Verify no extra fields
	if len(rawResponse) != len(requiredFields) {
		t.Errorf("expected %d fields, got %d", len(requiredFields), len(rawResponse))
	}
}
