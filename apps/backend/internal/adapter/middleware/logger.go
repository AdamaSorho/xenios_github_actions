package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// statusRecorder wraps http.ResponseWriter to capture the status code.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.statusCode = code
	sr.ResponseWriter.WriteHeader(code)
}

// Logger is middleware that logs each request in structured JSON format
// with request correlation IDs.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(recorder, r)

		duration := time.Since(start)
		entry := map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"status":     recorder.statusCode,
			"duration_ms": duration.Milliseconds(),
			"remote_addr": r.RemoteAddr,
		}

		reqID := GetRequestID(r.Context())
		if reqID != "" {
			entry["request_id"] = reqID
		}

		jsonBytes, err := json.Marshal(entry)
		if err != nil {
			log.Printf("failed to marshal log entry: %v", err)
			return
		}
		log.Println(string(jsonBytes))
	})
}
