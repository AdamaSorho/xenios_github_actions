package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenios/backend/internal/adapter/handler"
	"github.com/xenios/backend/internal/infrastructure/config"
)

func TestGetPort_Default(t *testing.T) {
	originalPort := os.Getenv("PORT")
	_ = os.Unsetenv("PORT")
	defer restoreEnv(t, "PORT", originalPort)

	port := getPort()
	if port != "8080" {
		t.Errorf("expected default port 8080, got %s", port)
	}
}

func TestGetPort_CustomPort(t *testing.T) {
	originalPort := os.Getenv("PORT")
	_ = os.Setenv("PORT", "9999")
	defer restoreEnv(t, "PORT", originalPort)

	port := getPort()
	if port != "9999" {
		t.Errorf("expected port 9999, got %s", port)
	}
}

func TestGetPort_EmptyString(t *testing.T) {
	originalPort := os.Getenv("PORT")
	_ = os.Setenv("PORT", "")
	defer restoreEnv(t, "PORT", originalPort)

	port := getPort()
	if port != "8080" {
		t.Errorf("expected default port 8080 for empty string, got %s", port)
	}
}

func TestSetupServer_Configuration(t *testing.T) {
	originalPort := os.Getenv("PORT")
	_ = os.Setenv("PORT", "8080")
	defer restoreEnv(t, "PORT", originalPort)

	server, cleanup := setupServer()
	defer cleanup()

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
	originalPort := os.Getenv("PORT")
	port := "18080"
	_ = os.Setenv("PORT", port)
	defer restoreEnv(t, "PORT", originalPort)

	server, cleanup := setupServer()
	defer cleanup()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Health endpoint is now at /api/health
	resp, err := http.Get("http://localhost:" + port + "/api/health")
	if err != nil {
		t.Fatalf("failed to call health endpoint: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("failed to shutdown server: %v", err)
	}
}

func TestSetupServer_VersionEndpoint(t *testing.T) {
	originalPort := os.Getenv("PORT")
	port := "18081"
	_ = os.Setenv("PORT", port)
	defer restoreEnv(t, "PORT", originalPort)

	server, cleanup := setupServer()
	defer cleanup()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://localhost:" + port + "/version")
	if err != nil {
		t.Fatalf("failed to call version endpoint: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("failed to shutdown server: %v", err)
	}
}

func TestSetupServer_CustomPort(t *testing.T) {
	originalPort := os.Getenv("PORT")
	_ = os.Setenv("PORT", "18082")
	defer restoreEnv(t, "PORT", originalPort)

	server, cleanup := setupServer()
	defer cleanup()

	if server.Addr != ":18082" {
		t.Errorf("expected Addr :18082, got %s", server.Addr)
	}
}

func TestSetupServer_NotNil(t *testing.T) {
	server, cleanup := setupServer()
	defer cleanup()

	if server == nil {
		t.Error("expected non-nil server")
	}
}

func TestSetupServer_HandlerNotNil(t *testing.T) {
	server, cleanup := setupServer()
	defer cleanup()

	if server.Handler == nil {
		t.Error("expected non-nil Handler")
	}
}

func TestSetupServer_TimeoutValues(t *testing.T) {
	server, cleanup := setupServer()
	defer cleanup()

	if server.ReadTimeout != 15*time.Second {
		t.Errorf("expected ReadTimeout 15s, got %v", server.ReadTimeout)
	}
	if server.WriteTimeout != 15*time.Second {
		t.Errorf("expected WriteTimeout 15s, got %v", server.WriteTimeout)
	}
	if server.IdleTimeout != 60*time.Second {
		t.Errorf("expected IdleTimeout 60s, got %v", server.IdleTimeout)
	}
}

func TestGetPort_VariousPortNumbers(t *testing.T) {
	testCases := []struct {
		name     string
		portEnv  string
		expected string
	}{
		{"Port 80", "80", "80"},
		{"Port 443", "443", "443"},
		{"Port 3000", "3000", "3000"},
		{"Port 8000", "8000", "8000"},
		{"Port 65535", "65535", "65535"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalPort := os.Getenv("PORT")
			_ = os.Setenv("PORT", tc.portEnv)
			defer restoreEnv(t, "PORT", originalPort)

			port := getPort()
			if port != tc.expected {
				t.Errorf("expected port %s, got %s", tc.expected, port)
			}
		})
	}
}

func TestSetupServer_BothEndpointsConfigured(t *testing.T) {
	originalPort := os.Getenv("PORT")
	port := "18083"
	_ = os.Setenv("PORT", port)
	defer restoreEnv(t, "PORT", originalPort)

	server, cleanup := setupServer()
	defer cleanup()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	healthResp, err := http.Get("http://localhost:" + port + "/api/health")
	if err != nil {
		t.Fatalf("failed to call health endpoint: %v", err)
	}
	defer func() { _ = healthResp.Body.Close() }()

	versionResp, err := http.Get("http://localhost:" + port + "/version")
	if err != nil {
		t.Fatalf("failed to call version endpoint: %v", err)
	}
	defer func() { _ = versionResp.Body.Close() }()

	if healthResp.StatusCode != http.StatusOK {
		t.Errorf("expected health status 200, got %d", healthResp.StatusCode)
	}
	if versionResp.StatusCode != http.StatusOK {
		t.Errorf("expected version status 200, got %d", versionResp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("failed to shutdown server: %v", err)
	}
}

func TestSetupHealthHandler_NoDatabaseURL(t *testing.T) {
	originalDBURL := os.Getenv("DATABASE_URL")
	_ = os.Unsetenv("DATABASE_URL")
	defer restoreEnv(t, "DATABASE_URL", originalDBURL)

	h, pool, cleanup := setupHealthHandler()
	defer cleanup()

	if h == nil {
		t.Error("expected non-nil handler when DATABASE_URL is not set")
	}
	if pool != nil {
		t.Error("expected nil pool when DATABASE_URL is not set")
	}
}

func TestSetupHealthHandler_InvalidDatabaseURL(t *testing.T) {
	originalDBURL := os.Getenv("DATABASE_URL")
	_ = os.Setenv("DATABASE_URL", "not-a-valid-url")
	defer restoreEnv(t, "DATABASE_URL", originalDBURL)

	h, pool, cleanup := setupHealthHandler()
	defer cleanup()

	if h == nil {
		t.Error("expected non-nil handler even with invalid DATABASE_URL")
	}
	if pool != nil {
		t.Error("expected nil pool with invalid DATABASE_URL")
	}
}

func TestSetupHealthHandler_UnreachableDatabase(t *testing.T) {
	originalDBURL := os.Getenv("DATABASE_URL")
	_ = os.Setenv("DATABASE_URL", "postgres://invalid:invalid@localhost:19999/nonexistent?connect_timeout=1")
	defer restoreEnv(t, "DATABASE_URL", originalDBURL)

	h, pool, cleanup := setupHealthHandler()
	defer cleanup()

	if h == nil {
		t.Error("expected non-nil handler even with unreachable database")
	}
	if pool != nil {
		t.Error("expected nil pool with unreachable database")
	}
}

func TestSetupHealthHandler_CleanupFunctionIsSafe(t *testing.T) {
	originalDBURL := os.Getenv("DATABASE_URL")
	_ = os.Unsetenv("DATABASE_URL")
	defer restoreEnv(t, "DATABASE_URL", originalDBURL)

	_, _, cleanup := setupHealthHandler()
	cleanup()
	cleanup() // Ensure idempotency
}

func TestRunServer_GracefulShutdown(t *testing.T) {
	originalPort := os.Getenv("PORT")
	_ = os.Setenv("PORT", "18084")
	defer restoreEnv(t, "PORT", originalPort)

	server, cleanup := setupServer()
	defer cleanup()

	signal.Reset(syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{})
	go func() {
		runServer(server)
		close(done)
	}()

	time.Sleep(200 * time.Millisecond)

	resp, err := http.Get("http://localhost:18084/api/health")
	if err != nil {
		t.Fatalf("server should be running: %v", err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find current process: %v", err)
	}
	_ = proc.Signal(syscall.SIGINT)

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("server did not shut down within timeout")
	}
}

func TestSetupServer_NoDatabaseURL_ReturnsLegacyHealth(t *testing.T) {
	originalPort := os.Getenv("PORT")
	originalDBURL := os.Getenv("DATABASE_URL")
	_ = os.Setenv("PORT", "18085")
	_ = os.Unsetenv("DATABASE_URL")
	defer func() {
		restoreEnv(t, "PORT", originalPort)
		restoreEnv(t, "DATABASE_URL", originalDBURL)
	}()

	server, cleanup := setupServer()
	defer cleanup()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://localhost:18085/api/health")
	if err != nil {
		t.Fatalf("failed to call health endpoint: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		t.Errorf("failed to shutdown server: %v", err)
	}
}

func TestSetupJobQueue_ReturnsHandlerAndWorker(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://test:test@localhost:19999/testdb")
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Close()

	queueHandler, w, jq := setupJobQueue(pool)

	if queueHandler == nil {
		t.Error("expected non-nil QueueHandler")
	}
	if w == nil {
		t.Error("expected non-nil Worker")
	}
	if jq == nil {
		t.Error("expected non-nil JobQueue")
	}
	if !w.IsRunning() {
		t.Error("expected worker to be running after setup")
	}

	w.Stop()
}

func TestSetupJobQueue_RegistersAllJobTypes(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://test:test@localhost:19999/testdb")
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Close()

	_, w, _ := setupJobQueue(pool)
	defer w.Stop()

	types := w.RegisteredJobTypes()
	if len(types) != 12 {
		t.Errorf("expected 12 registered job types, got %d", len(types))
	}
}

func TestSetupJobQueue_WorkerCanBeStopped(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://test:test@localhost:19999/testdb")
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Close()

	_, w, _ := setupJobQueue(pool)
	if !w.IsRunning() {
		t.Error("expected worker to be running")
	}

	w.Stop()
	if w.IsRunning() {
		t.Error("expected worker to be stopped")
	}
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

func TestConfigureRoutes_WithoutPool_NoJobEndpoints(t *testing.T) {
	cfg := &config.Config{
		Port:        "8080",
		CORSOrigins: []string{"http://localhost:3000"},
	}
	healthHandler := handler.NewHealthHandler()
	router, w, asyncAudit := configureRoutes(cfg, healthHandler, nil)
	defer asyncAudit.Stop()

	if router == nil {
		t.Error("expected non-nil router")
	}
	if w != nil {
		t.Error("expected nil worker when pool is nil")
	}
}

func TestConfigureRoutes_WithPool_RegistersJobEndpoints(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://test:test@localhost:19999/testdb")
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Close()

	cfg := &config.Config{
		Port:        "8080",
		CORSOrigins: []string{"http://localhost:3000"},
	}
	healthHandler := handler.NewHealthHandler()
	router, w, asyncAudit := configureRoutes(cfg, healthHandler, pool)
	defer asyncAudit.Stop()

	if router == nil {
		t.Error("expected non-nil router")
	}
	if w == nil {
		t.Error("expected non-nil worker when pool is provided")
	}
	if !w.IsRunning() {
		t.Error("expected worker to be running")
	}
	w.Stop()
}

func TestSetupServer_ProtectedRouteRequiresAuth(t *testing.T) {
	originalPort := os.Getenv("PORT")
	port := "18086"
	_ = os.Setenv("PORT", port)
	defer restoreEnv(t, "PORT", originalPort)

	server, cleanup := setupServer()
	defer cleanup()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Request to protected route without auth should return 401
	resp, err := http.Get("http://localhost:" + port + "/api/v1/coaches/coach-1/clients")
	if err != nil {
		t.Fatalf("failed to call protected endpoint: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body["code"] != "UNAUTHORIZED" {
		t.Errorf("expected code UNAUTHORIZED, got %v", body["code"])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("failed to shutdown server: %v", err)
	}
}

func TestSetupServer_CORSHeaders(t *testing.T) {
	originalPort := os.Getenv("PORT")
	port := "18087"
	_ = os.Setenv("PORT", port)
	defer restoreEnv(t, "PORT", originalPort)

	server, cleanup := setupServer()
	defer cleanup()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Send OPTIONS preflight request
	req, _ := http.NewRequest("OPTIONS", "http://localhost:"+port+"/api/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to send preflight request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// CORS should be handled
	allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://localhost:3000" {
		t.Errorf("expected Access-Control-Allow-Origin http://localhost:3000, got %s", allowOrigin)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("failed to shutdown server: %v", err)
	}
}

func TestSetupServer_RequestIDHeader(t *testing.T) {
	originalPort := os.Getenv("PORT")
	port := "18088"
	_ = os.Setenv("PORT", port)
	defer restoreEnv(t, "PORT", originalPort)

	server, cleanup := setupServer()
	defer cleanup()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://localhost:" + port + "/api/health")
	if err != nil {
		t.Fatalf("failed to call health endpoint: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	reqID := resp.Header.Get("X-Request-ID")
	if reqID == "" {
		t.Error("expected X-Request-ID header in response")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("failed to shutdown server: %v", err)
	}
}

func TestSetupServer_PanicRecovery(t *testing.T) {
	// Verify the middleware is applied by checking that the server doesn't crash
	// on a normal request (full panic recovery test is in middleware_test.go)
	originalPort := os.Getenv("PORT")
	port := "18089"
	_ = os.Setenv("PORT", port)
	defer restoreEnv(t, "PORT", originalPort)

	server, cleanup := setupServer()
	defer cleanup()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Normal request should work fine with recoverer middleware
	resp, err := http.Get("http://localhost:" + port + "/api/health")
	if err != nil {
		t.Fatalf("failed to call health endpoint: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("failed to shutdown server: %v", err)
	}
}

func restoreEnv(t *testing.T, key, val string) {
	t.Helper()
	if val != "" {
		_ = os.Setenv(key, val)
	} else {
		_ = os.Unsetenv(key)
	}
}
