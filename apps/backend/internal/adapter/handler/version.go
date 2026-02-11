package handler

import (
	"net/http"
	"os"
	"runtime"
)

// Version is the application version (will be wired to build flags later).
const Version = "0.1.0"

// VersionHandler handles HTTP requests for version information.
type VersionHandler struct{}

// NewVersionHandler creates a new VersionHandler.
func NewVersionHandler() *VersionHandler {
	return &VersionHandler{}
}

// VersionResponse is the JSON response format for version checks.
type VersionResponse struct {
	Version     string `json:"version"`
	Environment string `json:"environment"`
	GoVersion   string `json:"goVersion"`
}

// Version handles GET /version
func (h *VersionHandler) Version(w http.ResponseWriter, r *http.Request) {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}

	if err := respondJSON(w, http.StatusOK, VersionResponse{
		Version:     Version,
		Environment: environment,
		GoVersion:   runtime.Version(),
	}); err != nil {
		// Error already written to response, log it
		// TODO: Add proper logging when logger is available
		return
	}
}
