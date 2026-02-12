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
	S3          S3Config
}

// S3Config holds configuration for S3/R2 object storage.
type S3Config struct {
	Bucket    string
	Region    string
	Endpoint  string
	AccessKey string
	SecretKey string
}

// Load reads configuration from environment variables.
func Load() *Config {
	env := getEnvOrDefault("ENVIRONMENT", "development")
	return &Config{
		Port:        getEnvOrDefault("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		Environment: env,
		CORSOrigins: parseCORSOrigins(os.Getenv("CORS_ORIGINS")),
		S3: S3Config{
			Bucket:    getEnvOrDefault("S3_BUCKET", "xenios-"+env+"-uploads"),
			Region:    getEnvOrDefault("S3_REGION", "us-east-1"),
			Endpoint:  os.Getenv("S3_ENDPOINT"),
			AccessKey: os.Getenv("S3_ACCESS_KEY"),
			SecretKey: os.Getenv("S3_SECRET_KEY"),
		},
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
