package handler

import (
	"encoding/json"
	"net/http"
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
