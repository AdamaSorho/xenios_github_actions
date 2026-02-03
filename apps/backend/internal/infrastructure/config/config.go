package config

import "os"

// Config holds all configuration for the application.
type Config struct {
	Port        string
	DatabaseURL string
	Environment string
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
