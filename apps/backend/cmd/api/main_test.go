package main

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"
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
	server := setupServer()

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
	server := setupServer()

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
	server := setupServer()

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
	server := setupServer()

	// Assert
	expectedAddr := ":" + customPort
	if server.Addr != expectedAddr {
		t.Errorf("expected Addr %s, got %s", expectedAddr, server.Addr)
	}
}

func TestSetupServer_NotNil(t *testing.T) {
	// Act
	server := setupServer()

	// Assert
	if server == nil {
		t.Error("expected non-nil server")
	}
}

func TestSetupServer_HandlerNotNil(t *testing.T) {
	// Act
	server := setupServer()

	// Assert
	if server.Handler == nil {
		t.Error("expected non-nil Handler")
	}
}

func TestSetupServer_TimeoutValues(t *testing.T) {
	// Act
	server := setupServer()

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
	server := setupServer()

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
