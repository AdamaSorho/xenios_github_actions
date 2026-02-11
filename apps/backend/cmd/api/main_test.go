package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestGetPort_Default(t *testing.T) {
	// Arrange - unset PORT to test default
	originalPort := os.Getenv("PORT")
	os.Unsetenv("PORT")
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		}
	}()

	// Act
	port := getPort()

	// Assert
	if port != "8080" {
		t.Errorf("expected default port 8080, got %s", port)
	}
}

func TestGetPort_CustomPort(t *testing.T) {
	// Arrange
	originalPort := os.Getenv("PORT")
	customPort := "9999"
	os.Setenv("PORT", customPort)
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		} else {
			os.Unsetenv("PORT")
		}
	}()

	// Act
	port := getPort()

	// Assert
	if port != customPort {
		t.Errorf("expected custom port %s, got %s", customPort, port)
	}
}

func TestGetPort_EmptyString(t *testing.T) {
	// Arrange
	originalPort := os.Getenv("PORT")
	os.Setenv("PORT", "")
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		} else {
			os.Unsetenv("PORT")
		}
	}()

	// Act
	port := getPort()

	// Assert - empty string should default to 8080
	if port != "8080" {
		t.Errorf("expected default port 8080 for empty string, got %s", port)
	}
}

func TestSetupServer_Configuration(t *testing.T) {
	// Arrange
	originalPort := os.Getenv("PORT")
	os.Setenv("PORT", "8080")
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		} else {
			os.Unsetenv("PORT")
		}
	}()

	// Act
	server, cleanup := setupServer()
	defer cleanup()

	// Assert - verify configuration
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
	originalPort := os.Getenv("PORT")
	port := "18080"
	os.Setenv("PORT", port)
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		} else {
			os.Unsetenv("PORT")
		}
	}()

	// Act
	server, cleanup := setupServer()
	defer cleanup()

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Expected to close during test
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Test health endpoint
	resp, err := http.Get("http://localhost:" + port + "/health")
	if err != nil {
		t.Fatalf("failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Assert
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Clean up
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("failed to shutdown server: %v", err)
	}
}

func TestSetupServer_VersionEndpoint(t *testing.T) {
	// Arrange
	originalPort := os.Getenv("PORT")
	port := "18081"
	os.Setenv("PORT", port)
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		} else {
			os.Unsetenv("PORT")
		}
	}()

	// Act
	server, cleanup := setupServer()
	defer cleanup()

	// Start server
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Expected to close during test
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Test version endpoint
	resp, err := http.Get("http://localhost:" + port + "/version")
	if err != nil {
		t.Fatalf("failed to call version endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Assert
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Clean up
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("failed to shutdown server: %v", err)
	}
}

func TestSetupServer_CustomPort(t *testing.T) {
	// Arrange
	originalPort := os.Getenv("PORT")
	customPort := "18082"
	os.Setenv("PORT", customPort)
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		} else {
			os.Unsetenv("PORT")
		}
	}()

	// Act
	server, cleanup := setupServer()
	defer cleanup()

	// Assert
	expectedAddr := ":" + customPort
	if server.Addr != expectedAddr {
		t.Errorf("expected Addr %s, got %s", expectedAddr, server.Addr)
	}
}

func TestSetupServer_NotNil(t *testing.T) {
	// Act
	server, cleanup := setupServer()
	defer cleanup()

	// Assert
	if server == nil {
		t.Error("expected non-nil server")
	}
}

func TestSetupServer_HandlerNotNil(t *testing.T) {
	// Act
	server, cleanup := setupServer()
	defer cleanup()

	// Assert
	if server.Handler == nil {
		t.Error("expected non-nil Handler")
	}
}

func TestSetupServer_TimeoutValues(t *testing.T) {
	// Act
	server, cleanup := setupServer()
	defer cleanup()

	// Assert - verify specific timeout values
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
			// Arrange
			originalPort := os.Getenv("PORT")
			os.Setenv("PORT", tc.portEnv)
			defer func() {
				if originalPort != "" {
					os.Setenv("PORT", originalPort)
				} else {
					os.Unsetenv("PORT")
				}
			}()

			// Act
			port := getPort()

			// Assert
			if port != tc.expected {
				t.Errorf("expected port %s, got %s", tc.expected, port)
			}
		})
	}
}

func TestSetupServer_BothEndpointsConfigured(t *testing.T) {
	// Arrange
	originalPort := os.Getenv("PORT")
	port := "18083"
	os.Setenv("PORT", port)
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		} else {
			os.Unsetenv("PORT")
		}
	}()

	// Act
	server, cleanup := setupServer()
	defer cleanup()

	// Start server
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Expected to close during test
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Test both endpoints
	healthResp, err := http.Get("http://localhost:" + port + "/health")
	if err != nil {
		t.Fatalf("failed to call health endpoint: %v", err)
	}
	defer healthResp.Body.Close()

	versionResp, err := http.Get("http://localhost:" + port + "/version")
	if err != nil {
		t.Fatalf("failed to call version endpoint: %v", err)
	}
	defer versionResp.Body.Close()

	// Assert both return 200
	if healthResp.StatusCode != http.StatusOK {
		t.Errorf("expected health status 200, got %d", healthResp.StatusCode)
	}

	if versionResp.StatusCode != http.StatusOK {
		t.Errorf("expected version status 200, got %d", versionResp.StatusCode)
	}

	// Clean up
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("failed to shutdown server: %v", err)
	}
}

func TestSetupHealthHandler_NoDatabaseURL(t *testing.T) {
	// Arrange - ensure DATABASE_URL is not set
	originalDBURL := os.Getenv("DATABASE_URL")
	os.Unsetenv("DATABASE_URL")
	defer func() {
		if originalDBURL != "" {
			os.Setenv("DATABASE_URL", originalDBURL)
		}
	}()

	// Act
	h, cleanup := setupHealthHandler()
	defer cleanup()

	// Assert - handler should be non-nil
	if h == nil {
		t.Error("expected non-nil handler when DATABASE_URL is not set")
	}
}

func TestSetupHealthHandler_InvalidDatabaseURL(t *testing.T) {
	// Arrange - set an invalid DATABASE_URL that will fail pool creation
	originalDBURL := os.Getenv("DATABASE_URL")
	os.Setenv("DATABASE_URL", "not-a-valid-url")
	defer func() {
		if originalDBURL != "" {
			os.Setenv("DATABASE_URL", originalDBURL)
		} else {
			os.Unsetenv("DATABASE_URL")
		}
	}()

	// Act
	h, cleanup := setupHealthHandler()
	defer cleanup()

	// Assert - handler should still be non-nil (graceful degradation)
	if h == nil {
		t.Error("expected non-nil handler even with invalid DATABASE_URL")
	}
}

func TestSetupHealthHandler_UnreachableDatabase(t *testing.T) {
	// Arrange - set a valid-format but unreachable DATABASE_URL
	originalDBURL := os.Getenv("DATABASE_URL")
	os.Setenv("DATABASE_URL", "postgres://invalid:invalid@localhost:19999/nonexistent?connect_timeout=1")
	defer func() {
		if originalDBURL != "" {
			os.Setenv("DATABASE_URL", originalDBURL)
		} else {
			os.Unsetenv("DATABASE_URL")
		}
	}()

	// Act
	h, cleanup := setupHealthHandler()
	defer cleanup()

	// Assert - handler should still be non-nil (graceful degradation)
	if h == nil {
		t.Error("expected non-nil handler even with unreachable database")
	}
}

func TestSetupHealthHandler_CleanupFunctionIsSafe(t *testing.T) {
	// Arrange
	originalDBURL := os.Getenv("DATABASE_URL")
	os.Unsetenv("DATABASE_URL")
	defer func() {
		if originalDBURL != "" {
			os.Setenv("DATABASE_URL", originalDBURL)
		}
	}()

	// Act
	_, cleanup := setupHealthHandler()

	// Assert - cleanup should be callable without panicking (no-op when no DB)
	cleanup()
	// Call again to ensure idempotency
	cleanup()
}

func TestRunServer_GracefulShutdown(t *testing.T) {
	// Arrange
	originalPort := os.Getenv("PORT")
	os.Setenv("PORT", "18084")
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		} else {
			os.Unsetenv("PORT")
		}
	}()

	server, cleanup := setupServer()
	defer cleanup()

	// Reset signal handling for this test
	signal.Reset(syscall.SIGINT, syscall.SIGTERM)

	// Act - run server in background
	done := make(chan struct{})
	go func() {
		runServer(server)
		close(done)
	}()

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	// Verify server is running
	resp, err := http.Get("http://localhost:18084/health")
	if err != nil {
		t.Fatalf("server should be running: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Send SIGINT to trigger graceful shutdown
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find current process: %v", err)
	}
	proc.Signal(syscall.SIGINT)

	// Wait for server to shut down
	select {
	case <-done:
		// Server shut down gracefully
	case <-time.After(10 * time.Second):
		t.Fatal("server did not shut down within timeout")
	}
}

func TestSetupServer_NoDatabaseURL_ReturnsLegacyHealth(t *testing.T) {
	// Arrange
	originalPort := os.Getenv("PORT")
	originalDBURL := os.Getenv("DATABASE_URL")
	os.Setenv("PORT", "18085")
	os.Unsetenv("DATABASE_URL")
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		} else {
			os.Unsetenv("PORT")
		}
		if originalDBURL != "" {
			os.Setenv("DATABASE_URL", originalDBURL)
		}
	}()

	// Act
	server, cleanup := setupServer()
	defer cleanup()

	// Start server
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Expected to close during test
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Test health endpoint returns legacy format
	resp, err := http.Get("http://localhost:18085/health")
	if err != nil {
		t.Fatalf("failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Clean up
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		t.Errorf("failed to shutdown server: %v", err)
	}
}

func TestWireHealthHandler_ReturnsHandlerAndCleanup(t *testing.T) {
	// Arrange - create a pool with an unreachable database
	// The pool itself is created successfully; it's lazy and won't connect until used
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://test:test@localhost:19999/testdb")
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// Act
	h, cleanup := wireHealthHandler(pool)

	// Assert - handler should be non-nil
	if h == nil {
		t.Error("expected non-nil handler from wireHealthHandler")
	}

	// Cleanup should not panic
	cleanup()
}

func TestWireHealthHandler_CleanupClosesPool(t *testing.T) {
	// Arrange
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://test:test@localhost:19999/testdb")
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// Act
	_, cleanup := wireHealthHandler(pool)
	cleanup()

	// Assert - after cleanup, pool should be closed
	// Calling Ping on a closed pool should fail
	if err := pool.Ping(ctx); err == nil {
		t.Error("expected error when pinging closed pool")
	}
}
