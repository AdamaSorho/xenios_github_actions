package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type auditContextKey struct{}

// AuditContext holds request metadata for audit logging.
type AuditContext struct {
	IPAddress string
	UserAgent string
}

// AuditContextMiddleware extracts IP address and user agent from the request
// and stores them in the context for use by downstream audit logging.
func AuditContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ac := &AuditContext{
			IPAddress: extractIPAddress(r),
			UserAgent: r.UserAgent(),
		}
		ctx := context.WithValue(r.Context(), auditContextKey{}, ac)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetAuditContext retrieves the audit context from the request context.
func GetAuditContext(ctx context.Context) *AuditContext {
	ac, ok := ctx.Value(auditContextKey{}).(*AuditContext)
	if !ok {
		return nil
	}
	return ac
}

// extractIPAddress gets the client IP from request headers or RemoteAddr.
func extractIPAddress(r *http.Request) string {
	// Check X-Forwarded-For first (common with load balancers/proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain (original client)
		parts := strings.SplitN(xff, ",", 2)
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip
		}
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
