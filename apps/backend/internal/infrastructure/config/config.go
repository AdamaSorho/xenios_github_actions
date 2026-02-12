package config

import (
	"os"
	"strings"
)

// Config holds all configuration for the application.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	Environment string
	CORSOrigins []string
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Port:        getEnvOrDefault("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
		CORSOrigins: parseCORSOrigins(os.Getenv("CORS_ORIGINS")),
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func parseCORSOrigins(raw string) []string {
	defaults := []string{"http://localhost:3000", "http://localhost:8081"}
	if raw == "" {
		return defaults
	}
	var origins []string
	for _, s := range strings.Split(raw, ",") {
		if t := strings.TrimSpace(s); t != "" {
			origins = append(origins, t)
		}
	}
	if len(origins) == 0 {
		return defaults
	}
	return origins
}
