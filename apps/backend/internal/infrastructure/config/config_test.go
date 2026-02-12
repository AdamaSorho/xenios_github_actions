package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Arrange - clear env vars
	envVars := []string{"PORT", "DATABASE_URL", "JWT_SECRET", "ENVIRONMENT", "CORS_ORIGINS"}
	originals := make(map[string]string)
	for _, v := range envVars {
		originals[v] = os.Getenv(v)
		os.Unsetenv(v)
	}
	defer func() {
		for k, v := range originals {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	}()

	// Act
	cfg := Load()

	// Assert
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
		t.Errorf("expected environment development, got %s", cfg.Environment)
	}
	if len(cfg.CORSOrigins) != 2 {
		t.Errorf("expected 2 default CORS origins, got %d", len(cfg.CORSOrigins))
	}
}

func TestLoad_CustomValues(t *testing.T) {
	// Arrange
	envVars := map[string]string{
		"PORT":         "9090",
		"DATABASE_URL": "postgres://localhost/test",
		"JWT_SECRET":   "secret123",
		"ENVIRONMENT":  "production",
		"CORS_ORIGINS": "https://app.example.com,https://admin.example.com",
	}
	originals := make(map[string]string)
	for k, v := range envVars {
		originals[k] = os.Getenv(k)
		os.Setenv(k, v)
	}
	defer func() {
		for k, v := range originals {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	}()

	// Act
	cfg := Load()

	// Assert
	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://localhost/test" {
		t.Errorf("expected DatabaseURL postgres://localhost/test, got %s", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "secret123" {
		t.Errorf("expected JWTSecret secret123, got %s", cfg.JWTSecret)
	}
	if cfg.Environment != "production" {
		t.Errorf("expected environment production, got %s", cfg.Environment)
	}
	if len(cfg.CORSOrigins) != 2 {
		t.Fatalf("expected 2 CORS origins, got %d", len(cfg.CORSOrigins))
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
		t.Errorf("expected 2 default origins for empty string, got %d", len(origins))
	}
}

func TestParseCORSOrigins_SingleOrigin(t *testing.T) {
	origins := parseCORSOrigins("https://example.com")
	if len(origins) != 1 {
		t.Fatalf("expected 1 origin, got %d", len(origins))
	}
	if origins[0] != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", origins[0])
	}
}

func TestParseCORSOrigins_MultipleOrigins(t *testing.T) {
	origins := parseCORSOrigins("https://a.com, https://b.com , https://c.com")
	if len(origins) != 3 {
		t.Fatalf("expected 3 origins, got %d", len(origins))
	}
	expected := []string{"https://a.com", "https://b.com", "https://c.com"}
	for i, e := range expected {
		if origins[i] != e {
			t.Errorf("origin[%d]: expected %s, got %s", i, e, origins[i])
		}
	}
}

func TestParseCORSOrigins_WhitespaceOnly(t *testing.T) {
	origins := parseCORSOrigins("   ,  ,  ")
	// All entries are empty after trim, so should fall back to defaults
	if len(origins) != 2 {
		t.Errorf("expected 2 default origins for whitespace-only, got %d", len(origins))
	}
}

func TestGetEnvOrDefault_EnvSet(t *testing.T) {
	os.Setenv("TEST_CONFIG_VAR", "value")
	defer os.Unsetenv("TEST_CONFIG_VAR")

	result := getEnvOrDefault("TEST_CONFIG_VAR", "default")
	if result != "value" {
		t.Errorf("expected value, got %s", result)
	}
}

func TestGetEnvOrDefault_EnvNotSet(t *testing.T) {
	os.Unsetenv("TEST_CONFIG_VAR_UNSET")

	result := getEnvOrDefault("TEST_CONFIG_VAR_UNSET", "default")
	if result != "default" {
		t.Errorf("expected default, got %s", result)
	}
}

func TestGetEnvOrDefault_EnvEmpty(t *testing.T) {
	os.Setenv("TEST_CONFIG_VAR_EMPTY", "")
	defer os.Unsetenv("TEST_CONFIG_VAR_EMPTY")

	result := getEnvOrDefault("TEST_CONFIG_VAR_EMPTY", "default")
	if result != "default" {
		t.Errorf("expected default for empty env, got %s", result)
	}
}

func TestTrimSpace_Various(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"hello", "hello"},
		{"  ", ""},
		{"", ""},
		{"\thello\t", "hello"},
	}
	for _, tc := range tests {
		result := trimSpace(tc.input)
		if result != tc.expected {
			t.Errorf("trimSpace(%q): expected %q, got %q", tc.input, tc.expected, result)
		}
	}
}
