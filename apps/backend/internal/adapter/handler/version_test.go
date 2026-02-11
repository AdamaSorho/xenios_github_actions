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

// Boundary Tests

func TestVersionHandler_Version_EmptyStringEnvironment(t *testing.T) {
	// Arrange - test empty string vs unset
	originalEnv := os.Getenv("ENVIRONMENT")
	os.Setenv("ENVIRONMENT", "")
	defer os.Setenv("ENVIRONMENT", originalEnv)

	handler := NewVersionHandler()
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Version(rec, req)

	// Assert - empty string should default to "development"
	var response VersionResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expected := "development"
	if response.Environment != expected {
		t.Errorf("expected environment %q for empty string, got %q", expected, response.Environment)
	}
}

func TestVersionHandler_Version_SpecialCharactersInEnvironment(t *testing.T) {
	// Arrange
	originalEnv := os.Getenv("ENVIRONMENT")
	testCases := []string{
		"test-staging",
		"test_production",
		"test.env",
		"test@123",
		"test with spaces",
		"test\nwith\nnewlines",
		"test\twith\ttabs",
	}

	for _, testEnv := range testCases {
		t.Run(testEnv, func(t *testing.T) {
			os.Setenv("ENVIRONMENT", testEnv)
			defer os.Setenv("ENVIRONMENT", originalEnv)

			handler := NewVersionHandler()
			req := httptest.NewRequest(http.MethodGet, "/version", nil)
			rec := httptest.NewRecorder()

			// Act
			handler.Version(rec, req)

			// Assert - should return environment as-is
			var response VersionResponse
			if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if response.Environment != testEnv {
				t.Errorf("expected environment %q, got %q", testEnv, response.Environment)
			}
		})
	}
}

func TestVersionHandler_Version_ExtremelyLongEnvironment(t *testing.T) {
	// Arrange
	originalEnv := os.Getenv("ENVIRONMENT")
	longEnv := string(make([]byte, 10000)) // 10KB environment string
	for i := range longEnv {
		longEnv = longEnv[:i] + "a" + longEnv[i+1:]
	}
	os.Setenv("ENVIRONMENT", longEnv)
	defer os.Setenv("ENVIRONMENT", originalEnv)

	handler := NewVersionHandler()
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Version(rec, req)

	// Assert - should handle long strings
	var response VersionResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Environment != longEnv {
		t.Errorf("expected long environment string to be preserved")
	}
}

func TestVersionHandler_Version_UnicodeInEnvironment(t *testing.T) {
	// Arrange
	originalEnv := os.Getenv("ENVIRONMENT")
	testCases := []string{
		"test-环境",
		"test-окружение",
		"test-🚀",
		"test-émoji",
	}

	for _, testEnv := range testCases {
		t.Run(testEnv, func(t *testing.T) {
			os.Setenv("ENVIRONMENT", testEnv)
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

			if response.Environment != testEnv {
				t.Errorf("expected environment %q, got %q", testEnv, response.Environment)
			}
		})
	}
}

// Input Validation Tests

func TestVersionHandler_Version_POSTMethodNotAllowed(t *testing.T) {
	// Arrange
	handler := NewVersionHandler()
	req := httptest.NewRequest(http.MethodPost, "/version", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Version(rec, req)

	// Assert - currently accepts all methods, but handler still returns 200
	// This test documents current behavior
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestVersionHandler_Version_PUTMethod(t *testing.T) {
	// Arrange
	handler := NewVersionHandler()
	req := httptest.NewRequest(http.MethodPut, "/version", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Version(rec, req)

	// Assert - documents current behavior
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestVersionHandler_Version_DELETEMethod(t *testing.T) {
	// Arrange
	handler := NewVersionHandler()
	req := httptest.NewRequest(http.MethodDelete, "/version", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Version(rec, req)

	// Assert - documents current behavior
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestVersionHandler_Version_HEADMethod(t *testing.T) {
	// Arrange
	handler := NewVersionHandler()
	req := httptest.NewRequest(http.MethodHead, "/version", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Version(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Note: httptest.ResponseRecorder still captures the body for HEAD requests.
	// In a real http.Server, the body is stripped automatically.
	// We just verify the handler doesn't crash on HEAD requests.
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}
}

func TestVersionHandler_Version_OPTIONSMethod(t *testing.T) {
	// Arrange
	handler := NewVersionHandler()
	req := httptest.NewRequest(http.MethodOptions, "/version", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Version(rec, req)

	// Assert - documents current behavior
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestVersionHandler_Version_InvalidContentType(t *testing.T) {
	// Arrange
	handler := NewVersionHandler()
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	// Act
	handler.Version(rec, req)

	// Assert - should still work, GET requests ignore Content-Type
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}
}

func TestVersionHandler_Version_WithQueryParameters(t *testing.T) {
	// Arrange
	handler := NewVersionHandler()
	req := httptest.NewRequest(http.MethodGet, "/version?foo=bar&baz=qux", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.Version(rec, req)

	// Assert - should ignore query parameters
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response VersionResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Version == "" {
		t.Error("expected version to be non-empty")
	}
}

func TestVersionHandler_Version_WithCustomHeaders(t *testing.T) {
	// Arrange
	handler := NewVersionHandler()
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	req.Header.Set("X-Custom-Header", "test-value")
	req.Header.Set("User-Agent", "test-agent")
	rec := httptest.NewRecorder()

	// Act
	handler.Version(rec, req)

	// Assert - should work with custom headers
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestVersionHandler_NewVersionHandler_NotNil(t *testing.T) {
	// Act
	handler := NewVersionHandler()

	// Assert
	if handler == nil {
		t.Error("expected non-nil handler")
	}
}

func TestVersionHandler_Version_ConcurrentRequests(t *testing.T) {
	// Arrange
	handler := NewVersionHandler()
	numRequests := 100

	// Act - fire concurrent requests
	done := make(chan bool, numRequests)
	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/version", nil)
			rec := httptest.NewRecorder()
			handler.Version(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
			}
			done <- true
		}()
	}

	// Assert - wait for all to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
}
