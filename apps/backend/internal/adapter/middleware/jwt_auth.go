package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type userClaimsKey struct{}

// UserClaims holds the parsed JWT claims available in the request context.
type UserClaims struct {
	Subject string
}

// JWTAuth returns middleware that validates JWT Bearer tokens.
// In production (non-empty secret), tokens are cryptographically verified.
// The parsed subject claim is stored in the request context.
func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondUnauthorized(w, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				respondUnauthorized(w, "invalid authorization format")
				return
			}

			tokenStr := strings.TrimSpace(parts[1])
			if tokenStr == "" {
				respondUnauthorized(w, "empty token")
				return
			}

			if secret == "" {
				// No secret configured — accept token without verification (development only).
				// Extract claims without validation.
				parser := jwt.NewParser()
				token, _, err := parser.ParseUnverified(tokenStr, jwt.MapClaims{})
				if err != nil {
					respondUnauthorized(w, "invalid token")
					return
				}
				claims, ok := token.Claims.(jwt.MapClaims)
				if !ok {
					respondUnauthorized(w, "invalid token claims")
					return
				}
				sub, _ := claims.GetSubject()
				ctx := context.WithValue(r.Context(), userClaimsKey{}, &UserClaims{Subject: sub})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Verify token with HMAC secret
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				respondUnauthorized(w, "invalid or expired token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				respondUnauthorized(w, "invalid token claims")
				return
			}

			sub, _ := claims.GetSubject()
			ctx := context.WithValue(r.Context(), userClaimsKey{}, &UserClaims{Subject: sub})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserClaims retrieves parsed JWT claims from the request context.
func GetUserClaims(ctx context.Context) *UserClaims {
	claims, ok := ctx.Value(userClaimsKey{}).(*UserClaims)
	if !ok {
		return nil
	}
	return claims
}

func respondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
		"code":  "UNAUTHORIZED",
	})
}
