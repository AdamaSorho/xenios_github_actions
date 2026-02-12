package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuditContextMiddleware_SetsIPAndUserAgent(t *testing.T) {
	var captured *AuditContext

	handler := AuditContextMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = GetAuditContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("User-Agent", "TestAgent/1.0")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if captured == nil {
		t.Fatal("expected audit context to be set")
	}
	if captured.IPAddress != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got '%s'", captured.IPAddress)
	}
	if captured.UserAgent != "TestAgent/1.0" {
		t.Errorf("expected User-Agent 'TestAgent/1.0', got '%s'", captured.UserAgent)
	}
}

func TestAuditContextMiddleware_XForwardedFor_UsesFirstIP(t *testing.T) {
	var captured *AuditContext

	handler := AuditContextMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = GetAuditContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18, 150.172.238.178")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if captured == nil {
		t.Fatal("expected audit context")
	}
	if captured.IPAddress != "203.0.113.50" {
		t.Errorf("expected IP '203.0.113.50', got '%s'", captured.IPAddress)
	}
}

func TestAuditContextMiddleware_XRealIP_Used(t *testing.T) {
	var captured *AuditContext

	handler := AuditContextMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = GetAuditContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Real-IP", "198.51.100.42")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if captured == nil {
		t.Fatal("expected audit context")
	}
	if captured.IPAddress != "198.51.100.42" {
		t.Errorf("expected IP '198.51.100.42', got '%s'", captured.IPAddress)
	}
}

func TestGetAuditContext_NoContext_ReturnsNil(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	ac := GetAuditContext(req.Context())
	if ac != nil {
		t.Error("expected nil audit context when middleware not applied")
	}
}
