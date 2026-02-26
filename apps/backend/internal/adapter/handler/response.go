package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/xenios/backend/internal/adapter/middleware"
	"github.com/xenios/backend/internal/usecase"
)

func respondJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}
	return nil
}

func respondError(w http.ResponseWriter, status int, message string) {
	_ = respondJSON(w, status, ErrorResponse{Error: message})
}

func respondErrorWithCode(w http.ResponseWriter, status int, message, code string, details map[string]interface{}) {
	resp := ErrorResponse{
		Error: message,
		Code:  code,
	}
	if details != nil {
		resp.Details = details
	}
	_ = respondJSON(w, status, resp)
}

// ErrorResponse is the standardized JSON error response format.
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// requireAuth extracts user claims from the request context.
// Returns nil and writes a 401 response if authentication is missing.
func requireAuth(w http.ResponseWriter, r *http.Request) *middleware.UserClaims {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return nil
	}
	return claims
}

// requireURLParam extracts a URL parameter by name.
// Returns empty string and writes a 400 response if the parameter is missing.
func requireURLParam(w http.ResponseWriter, r *http.Request, param, label string) string {
	val := chi.URLParam(r, param)
	if val == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing "+label, "INVALID_REQUEST", nil)
		return ""
	}
	return val
}

// decodeJSON decodes a JSON request body into the provided destination.
// Returns false and writes a 400 response if the JSON is invalid.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		respondErrorWithCode(w, http.StatusBadRequest, "invalid JSON body", "INVALID_JSON", nil)
		return false
	}
	return true
}

// handleUseCaseError maps use case errors to appropriate HTTP responses.
// Returns true if an error was handled (and a response was written).
func handleUseCaseError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if usecase.IsValidationError(err) {
		respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
		return true
	}
	if usecase.IsAuthenticationError(err) {
		respondErrorWithCode(w, http.StatusForbidden, err.Error(), "FORBIDDEN", nil)
		return true
	}
	respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
	return true
}
