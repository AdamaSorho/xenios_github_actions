package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_GeneratesID(t *testing.T) {
	// Arrange
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r.Context())
		if id == "" {
			t.Error("expected request ID to be set in context")
		}
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	responseID := rec.Header().Get("X-Request-ID")
	if responseID == "" {
		t.Error("expected X-Request-ID header to be set")
	}
	if len(responseID) != 32 { // 16 bytes = 32 hex chars
		t.Errorf("expected 32-char hex ID, got %d chars: %s", len(responseID), responseID)
	}
}

func TestRequestID_PreservesExistingID(t *testing.T) {
	// Arrange
	existingID := "my-custom-request-id"
	var capturedID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	if capturedID != existingID {
		t.Errorf("expected context ID %s, got %s", existingID, capturedID)
	}
	if rec.Header().Get("X-Request-ID") != existingID {
		t.Errorf("expected response header %s, got %s", existingID, rec.Header().Get("X-Request-ID"))
	}
}

func TestRequestID_UniquePerRequest(t *testing.T) {
	// Arrange
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Act
		handler.ServeHTTP(rec, req)

		id := rec.Header().Get("X-Request-ID")
		if ids[id] {
			t.Errorf("duplicate request ID generated: %s", id)
		}
		ids[id] = true
	}
}

func TestGetRequestID_NoValue(t *testing.T) {
	ctx := context.Background()
	id := GetRequestID(ctx)
	if id != "" {
		t.Errorf("expected empty string for context without request ID, got %s", id)
	}
}

func TestGetRequestID_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), requestIDKey, 12345)
	id := GetRequestID(ctx)
	if id != "" {
		t.Errorf("expected empty string for non-string value, got %s", id)
	}
}
