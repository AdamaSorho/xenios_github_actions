package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	originalPort := os.Getenv("PORT")
	originalDBURL := os.Getenv("DATABASE_URL")
	originalJWT := os.Getenv("JWT_SECRET")
	originalEnv := os.Getenv("ENVIRONMENT")
	originalCORS := os.Getenv("CORS_ORIGINS")
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("CORS_ORIGINS")
	defer func() {
		restoreEnv("PORT", originalPort)
		restoreEnv("DATABASE_URL", originalDBURL)
		restoreEnv("JWT_SECRET", originalJWT)
		restoreEnv("ENVIRONMENT", originalEnv)
		restoreEnv("CORS_ORIGINS", originalCORS)
	}()

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.DatabaseURL != "" {
		t.Errorf("expected empty DatabaseURL, got %s", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "" {
		t.Errorf("expected empty JWTSecret, got %s", cfg.JWTSecret)
	}
	if cfg.Environment != "development" {
		t.Errorf("expected development environment, got %s", cfg.Environment)
	}
	if len(cfg.CORSOrigins) != 2 {
		t.Errorf("expected 2 default CORS origins, got %d", len(cfg.CORSOrigins))
	}
	if cfg.CORSOrigins[0] != "http://localhost:3000" {
		t.Errorf("expected first CORS origin http://localhost:3000, got %s", cfg.CORSOrigins[0])
	}
	if cfg.CORSOrigins[1] != "http://localhost:8081" {
		t.Errorf("expected second CORS origin http://localhost:8081, got %s", cfg.CORSOrigins[1])
	}
}

func TestLoad_CustomValues(t *testing.T) {
	originalPort := os.Getenv("PORT")
	originalDBURL := os.Getenv("DATABASE_URL")
	originalJWT := os.Getenv("JWT_SECRET")
	originalEnv := os.Getenv("ENVIRONMENT")
	originalCORS := os.Getenv("CORS_ORIGINS")
	os.Setenv("PORT", "9090")
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("JWT_SECRET", "my-secret")
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("CORS_ORIGINS", "https://app.example.com, https://admin.example.com")
	defer func() {
		restoreEnv("PORT", originalPort)
		restoreEnv("DATABASE_URL", originalDBURL)
		restoreEnv("JWT_SECRET", originalJWT)
		restoreEnv("ENVIRONMENT", originalEnv)
		restoreEnv("CORS_ORIGINS", originalCORS)
	}()

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://localhost/test" {
		t.Errorf("expected DatabaseURL postgres://localhost/test, got %s", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "my-secret" {
		t.Errorf("expected JWTSecret my-secret, got %s", cfg.JWTSecret)
	}
	if cfg.Environment != "production" {
		t.Errorf("expected production environment, got %s", cfg.Environment)
	}
	if len(cfg.CORSOrigins) != 2 {
		t.Errorf("expected 2 CORS origins, got %d", len(cfg.CORSOrigins))
	}
	if cfg.CORSOrigins[0] != "https://app.example.com" {
		t.Errorf("expected first origin https://app.example.com, got %s", cfg.CORSOrigins[0])
	}
	if cfg.CORSOrigins[1] != "https://admin.example.com" {
		t.Errorf("expected second origin https://admin.example.com, got %s", cfg.CORSOrigins[1])
	}
}

func TestParseCORSOrigins_Empty(t *testing.T) {
	origins := parseCORSOrigins("")
	if len(origins) != 2 {
		t.Errorf("expected 2 defaults, got %d", len(origins))
	}
}

func TestParseCORSOrigins_SingleOrigin(t *testing.T) {
	origins := parseCORSOrigins("https://example.com")
	if len(origins) != 1 {
		t.Errorf("expected 1 origin, got %d", len(origins))
	}
	if origins[0] != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", origins[0])
	}
}

func TestParseCORSOrigins_WithWhitespace(t *testing.T) {
	origins := parseCORSOrigins("  https://a.com , https://b.com  ,  ")
	if len(origins) != 2 {
		t.Errorf("expected 2 origins, got %d", len(origins))
	}
	if origins[0] != "https://a.com" {
		t.Errorf("expected https://a.com, got %s", origins[0])
	}
	if origins[1] != "https://b.com" {
		t.Errorf("expected https://b.com, got %s", origins[1])
	}
}

func TestParseCORSOrigins_AllEmpty(t *testing.T) {
	origins := parseCORSOrigins(",,,")
	if len(origins) != 2 {
		t.Errorf("expected 2 defaults for all-empty input, got %d", len(origins))
	}
}

func TestGetEnvOrDefault_WithValue(t *testing.T) {
	os.Setenv("TEST_CONFIG_KEY", "test-value")
	defer os.Unsetenv("TEST_CONFIG_KEY")

	val := getEnvOrDefault("TEST_CONFIG_KEY", "default")
	if val != "test-value" {
		t.Errorf("expected test-value, got %s", val)
	}
}

func TestGetEnvOrDefault_WithDefault(t *testing.T) {
	os.Unsetenv("TEST_CONFIG_MISSING")

	val := getEnvOrDefault("TEST_CONFIG_MISSING", "default-val")
	if val != "default-val" {
		t.Errorf("expected default-val, got %s", val)
	}
}

func restoreEnv(key, val string) {
	if val != "" {
		os.Setenv(key, val)
	} else {
		os.Unsetenv(key)
	}
}
