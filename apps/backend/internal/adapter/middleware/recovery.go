package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
)

// Recoverer is a middleware that recovers from panics, logs the panic, and returns
// a 500 Internal Server Error with a standard JSON error response.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				reqID := GetRequestID(r.Context())
				log.Printf("[PANIC] request_id=%s panic=%v stack=%s", reqID, rec, debug.Stack())

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "internal server error",
					"code":  "INTERNAL_ERROR",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
