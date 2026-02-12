package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLogger_LogsRequestInfo(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := Logger(inner)
	req := httptest.NewRequest("GET", "/test-path", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := buf.String()
	if !strings.Contains(output, `"method":"GET"`) {
		t.Errorf("expected method in log output, got: %s", output)
	}
	if !strings.Contains(output, `"path":"/test-path"`) {
		t.Errorf("expected path in log output, got: %s", output)
	}
	if !strings.Contains(output, `"status":200`) {
		t.Errorf("expected status 200 in log output, got: %s", output)
	}
	if !strings.Contains(output, `"duration_ms"`) {
		t.Errorf("expected duration_ms in log output, got: %s", output)
	}
}

func TestLogger_CapturesStatusCode(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	handler := Logger(inner)
	req := httptest.NewRequest("GET", "/missing", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := buf.String()
	if !strings.Contains(output, `"status":404`) {
		t.Errorf("expected status 404 in log output, got: %s", output)
	}
}

func TestLogger_IncludesRequestID(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := Logger(inner)
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), requestIDKey, "test-req-123")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := buf.String()
	if !strings.Contains(output, `"request_id":"test-req-123"`) {
		t.Errorf("expected request_id in log output, got: %s", output)
	}
}

func TestLogger_DefaultStatus200(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write without calling WriteHeader — should default to 200
		_, _ = w.Write([]byte("ok"))
	})

	handler := Logger(inner)
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := buf.String()
	if !strings.Contains(output, `"status":200`) {
		t.Errorf("expected default status 200, got: %s", output)
	}
}

func TestLogger_OutputIsValidJSON(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := Logger(inner)
	req := httptest.NewRequest("POST", "/api/v1/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Extract JSON from log line (skip timestamp prefix)
	output := buf.String()
	// Find the JSON object in the log output
	jsonStart := strings.Index(output, "{")
	if jsonStart < 0 {
		t.Fatalf("no JSON found in log output: %s", output)
	}
	jsonStr := strings.TrimSpace(output[jsonStart:])

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Errorf("log output is not valid JSON: %v, output: %s", err, jsonStr)
	}
}

func TestStatusRecorder_WriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	sr := &statusRecorder{ResponseWriter: rec, statusCode: http.StatusOK}

	sr.WriteHeader(http.StatusCreated)

	if sr.statusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", sr.statusCode)
	}
}
