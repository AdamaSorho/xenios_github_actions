package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestLogger_LogsRequest(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	handler := RequestLogger(&buf)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	var entry LogEntry
	if err := json.NewDecoder(&buf).Decode(&entry); err != nil {
		t.Fatalf("failed to decode log entry: %v", err)
	}

	if entry.Method != "GET" {
		t.Errorf("expected method GET, got %s", entry.Method)
	}
	if entry.Path != "/api/v1/health" {
		t.Errorf("expected path /api/v1/health, got %s", entry.Path)
	}
	if entry.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", entry.StatusCode)
	}
	if entry.Timestamp == "" {
		t.Error("expected timestamp to be set")
	}
	if entry.Duration == "" {
		t.Error("expected duration to be set")
	}
}

func TestRequestLogger_LogsStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"OK", http.StatusOK},
		{"NotFound", http.StatusNotFound},
		{"InternalError", http.StatusInternalServerError},
		{"Created", http.StatusCreated},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			handler := RequestLogger(&buf)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			}))
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			var entry LogEntry
			if err := json.NewDecoder(&buf).Decode(&entry); err != nil {
				t.Fatalf("failed to decode log entry: %v", err)
			}

			if entry.StatusCode != tc.statusCode {
				t.Errorf("expected status %d, got %d", tc.statusCode, entry.StatusCode)
			}
		})
	}
}

func TestRequestLogger_IncludesRequestID(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	handler := RequestLogger(&buf)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), requestIDKey, "test-req-id")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	var entry LogEntry
	if err := json.NewDecoder(&buf).Decode(&entry); err != nil {
		t.Fatalf("failed to decode log entry: %v", err)
	}

	if entry.RequestID != "test-req-id" {
		t.Errorf("expected request ID test-req-id, got %s", entry.RequestID)
	}
}

func TestRequestLogger_DefaultStatusOK(t *testing.T) {
	// Arrange - handler that writes body without calling WriteHeader
	var buf bytes.Buffer
	handler := RequestLogger(&buf)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	var entry LogEntry
	if err := json.NewDecoder(&buf).Decode(&entry); err != nil {
		t.Fatalf("failed to decode log entry: %v", err)
	}

	if entry.StatusCode != http.StatusOK {
		t.Errorf("expected default status 200, got %d", entry.StatusCode)
	}
}

func TestStatusRecorder_CapturesWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	sr := &statusRecorder{ResponseWriter: rec, statusCode: http.StatusOK}

	sr.WriteHeader(http.StatusNotFound)

	if sr.statusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", sr.statusCode)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected underlying recorder status 404, got %d", rec.Code)
	}
}
