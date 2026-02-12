package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// LogEntry represents a structured JSON log entry for an HTTP request.
type LogEntry struct {
	Timestamp  string `json:"timestamp"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	StatusCode int    `json:"status_code"`
	Duration   string `json:"duration"`
	RequestID  string `json:"request_id,omitempty"`
	RemoteAddr string `json:"remote_addr"`
}

// statusRecorder wraps http.ResponseWriter to capture the status code.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.statusCode = code
	sr.ResponseWriter.WriteHeader(code)
}

// RequestLogger returns middleware that logs each request as structured JSON.
func RequestLogger(output io.Writer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rec, r)

			entry := LogEntry{
				Timestamp:  start.UTC().Format(time.RFC3339),
				Method:     r.Method,
				Path:       r.URL.Path,
				StatusCode: rec.statusCode,
				Duration:   time.Since(start).String(),
				RequestID:  GetRequestID(r.Context()),
				RemoteAddr: r.RemoteAddr,
			}

			json.NewEncoder(output).Encode(entry)
		})
	}
}
