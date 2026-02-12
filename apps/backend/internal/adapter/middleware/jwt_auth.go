package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

// JWTAuth returns middleware that validates the presence of a Bearer token
// in the Authorization header. For now, it only checks that a token is present.
// Full JWT validation (signature verification, claims) will be added when the
// auth feature is implemented.
func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondUnauthorized(w, "missing authorization header")
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				respondUnauthorized(w, "invalid authorization format, expected Bearer token")
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				respondUnauthorized(w, "empty bearer token")
				return
			}

			// TODO: Implement full JWT validation (signature, claims, expiry)
			// For now, just verify a token is present
			next.ServeHTTP(w, r)
		})
	}
}

func respondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
		"code":  "UNAUTHORIZED",
	})
}
