package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenios/backend/internal/infrastructure/config"
)

func testConfig(port string) *config.Config {
	return &config.Config{
		Port:        port,
		Environment: "test",
		CORSOrigins: []string{"http://localhost:3000"},
	}
}

func TestSetupServer_Configuration(t *testing.T) {
	// Arrange
	cfg := testConfig("8080")

	// Act
	server, cleanup := setupServer(cfg)
	defer cleanup()

	// Assert
	if server.Addr != ":8080" {
		t.Errorf("expected Addr :8080, got %s", server.Addr)
	}
	if server.ReadTimeout != 15*time.Second {
		t.Errorf("expected ReadTimeout 15s, got %v", server.ReadTimeout)
	}
	if server.WriteTimeout != 15*time.Second {
		t.Errorf("expected WriteTimeout 15s, got %v", server.WriteTimeout)
	}
	if server.IdleTimeout != 60*time.Second {
		t.Errorf("expected IdleTimeout 60s, got %v", server.IdleTimeout)
	}
	if server.Handler == nil {
		t.Error("expected non-nil Handler")
	}
}

func TestSetupServer_HealthEndpoint(t *testing.T) {
	// Arrange
	cfg := testConfig("18080")
	server, cleanup := setupServer(cfg)
	defer cleanup()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Expected
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Act
	resp, err := http.Get("http://localhost:18080/health")
	if err != nil {
		t.Fatalf("failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Assert
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify X-Request-ID header is set (middleware chain working)
	if resp.Header.Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header from middleware")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

func TestSetupServer_VersionEndpoint(t *testing.T) {
	// Arrange
	cfg := testConfig("18081")
	server, cleanup := setupServer(cfg)
	defer cleanup()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Expected
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Act
	resp, err := http.Get("http://localhost:18081/version")
	if err != nil {
		t.Fatalf("failed to call version endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Assert
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

func TestSetupServer_ProtectedEndpoint_NoAuth(t *testing.T) {
	// Arrange
	cfg := testConfig("18082")
	server, cleanup := setupServer(cfg)
	defer cleanup()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Expected
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Act - call protected endpoint without JWT
	resp, err := http.Get("http://localhost:18082/api/v1/coaches/coach-1/clients")
	if err != nil {
		t.Fatalf("failed to call protected endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Assert - should get 401
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["code"] != "UNAUTHORIZED" {
		t.Errorf("expected code UNAUTHORIZED, got %v", response["code"])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

func TestSetupServer_ProtectedEndpoint_WithAuth(t *testing.T) {
	// Arrange
	cfg := testConfig("18083")
	server, cleanup := setupServer(cfg)
	defer cleanup()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Expected
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Act - call protected endpoint with Bearer token
	req, _ := http.NewRequest(http.MethodGet, "http://localhost:18083/api/v1/coaches/coach-1/clients", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to call protected endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Assert - should get 200 with empty list
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

func TestSetupServer_CoachClientCreate(t *testing.T) {
	// Arrange
	cfg := testConfig("18084")
	server, cleanup := setupServer(cfg)
	defer cleanup()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Expected
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Act - create a coach-client relationship
	body := strings.NewReader(`{"client_id":"client-1"}`)
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:18084/api/v1/coaches/coach-1/clients", body)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to call create endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Assert
	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 201, got %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'data' field in response")
	}
	if data["coach_id"] != "coach-1" {
		t.Errorf("expected coach_id 'coach-1', got %v", data["coach_id"])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

func TestSetupServer_CustomPort(t *testing.T) {
	// Arrange
	cfg := testConfig("18086")

	// Act
	server, cleanup := setupServer(cfg)
	defer cleanup()

	// Assert
	if server.Addr != ":18086" {
		t.Errorf("expected Addr :18086, got %s", server.Addr)
	}
}

func TestSetupServer_NotNil(t *testing.T) {
	cfg := testConfig("8080")
	server, cleanup := setupServer(cfg)
	defer cleanup()
	if server == nil {
		t.Error("expected non-nil server")
	}
}

func TestSetupServer_HandlerNotNil(t *testing.T) {
	cfg := testConfig("8080")
	server, cleanup := setupServer(cfg)
	defer cleanup()
	if server.Handler == nil {
		t.Error("expected non-nil Handler")
	}
}

func TestSetupHealthHandler_NoDatabaseURL(t *testing.T) {
	originalDBURL := os.Getenv("DATABASE_URL")
	os.Unsetenv("DATABASE_URL")
	defer func() {
		if originalDBURL != "" {
			os.Setenv("DATABASE_URL", originalDBURL)
		}
	}()

	h, cleanup := setupHealthHandler()
	defer cleanup()

	if h == nil {
		t.Error("expected non-nil handler when DATABASE_URL is not set")
	}
}

func TestSetupHealthHandler_InvalidDatabaseURL(t *testing.T) {
	originalDBURL := os.Getenv("DATABASE_URL")
	os.Setenv("DATABASE_URL", "not-a-valid-url")
	defer func() {
		if originalDBURL != "" {
			os.Setenv("DATABASE_URL", originalDBURL)
		} else {
			os.Unsetenv("DATABASE_URL")
		}
	}()

	h, cleanup := setupHealthHandler()
	defer cleanup()

	if h == nil {
		t.Error("expected non-nil handler even with invalid DATABASE_URL")
	}
}

func TestSetupHealthHandler_UnreachableDatabase(t *testing.T) {
	originalDBURL := os.Getenv("DATABASE_URL")
	os.Setenv("DATABASE_URL", "postgres://invalid:invalid@localhost:19999/nonexistent?connect_timeout=1")
	defer func() {
		if originalDBURL != "" {
			os.Setenv("DATABASE_URL", originalDBURL)
		} else {
			os.Unsetenv("DATABASE_URL")
		}
	}()

	h, cleanup := setupHealthHandler()
	defer cleanup()

	if h == nil {
		t.Error("expected non-nil handler even with unreachable database")
	}
}

func TestSetupHealthHandler_CleanupFunctionIsSafe(t *testing.T) {
	originalDBURL := os.Getenv("DATABASE_URL")
	os.Unsetenv("DATABASE_URL")
	defer func() {
		if originalDBURL != "" {
			os.Setenv("DATABASE_URL", originalDBURL)
		}
	}()

	_, cleanup := setupHealthHandler()
	cleanup()
	cleanup() // idempotent
}

func TestRunServer_GracefulShutdown(t *testing.T) {
	// Arrange
	cfg := testConfig("18087")
	server, cleanup := setupServer(cfg)
	defer cleanup()

	signal.Reset(syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{})
	go func() {
		runServer(server)
		close(done)
	}()

	time.Sleep(200 * time.Millisecond)

	resp, err := http.Get("http://localhost:18087/health")
	if err != nil {
		t.Fatalf("server should be running: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find current process: %v", err)
	}
	proc.Signal(syscall.SIGINT)

	select {
	case <-done:
		// Server shut down gracefully
	case <-time.After(10 * time.Second):
		t.Fatal("server did not shut down within timeout")
	}
}

func TestSetupServer_CORSHeaders(t *testing.T) {
	// Arrange
	cfg := testConfig("18088")
	cfg.CORSOrigins = []string{"http://localhost:3000"}
	server, cleanup := setupServer(cfg)
	defer cleanup()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Expected
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Act - send preflight OPTIONS request
	req, _ := http.NewRequest(http.MethodOptions, "http://localhost:18088/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to send OPTIONS request: %v", err)
	}
	defer resp.Body.Close()

	// Assert - CORS headers should be present
	if resp.Header.Get("Access-Control-Allow-Origin") == "" {
		t.Error("expected Access-Control-Allow-Origin header")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

func TestWireHealthHandler_ReturnsHandlerAndCleanup(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://test:test@localhost:19999/testdb")
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	h, cleanup := wireHealthHandler(pool)
	if h == nil {
		t.Error("expected non-nil handler from wireHealthHandler")
	}
	cleanup()
}

func TestWireHealthHandler_CleanupClosesPool(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://test:test@localhost:19999/testdb")
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	_, cleanup := wireHealthHandler(pool)
	cleanup()

	if err := pool.Ping(ctx); err == nil {
		t.Error("expected error when pinging closed pool")
	}
}

func TestSetupServer_InvalidJSONBody(t *testing.T) {
	// Arrange
	cfg := testConfig("18089")
	server, cleanup := setupServer(cfg)
	defer cleanup()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Expected
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Act - send invalid JSON to a POST endpoint
	body := strings.NewReader("not json {{{")
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:18089/api/v1/coaches/coach-1/clients", body)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Assert - should get 400 with descriptive error
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["code"] != "INVALID_JSON" {
		t.Errorf("expected code 'INVALID_JSON', got %v", response["code"])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
