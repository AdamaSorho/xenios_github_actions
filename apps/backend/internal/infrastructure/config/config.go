package config

import "os"

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	Environment string
	CORSOrigins []string
}

// Load reads configuration from environment variables with sensible defaults.
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
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func parseCORSOrigins(raw string) []string {
	if raw == "" {
		return []string{"http://localhost:3000", "http://localhost:8081"}
	}
	var origins []string
	start := 0
	for i := 0; i <= len(raw); i++ {
		if i == len(raw) || raw[i] == ',' {
			s := trimSpace(raw[start:i])
			if s != "" {
				origins = append(origins, s)
			}
			start = i + 1
		}
	}
	if len(origins) == 0 {
		return []string{"http://localhost:3000", "http://localhost:8081"}
	}
	return origins
}

func trimSpace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
