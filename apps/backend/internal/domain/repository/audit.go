package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// AuditRepository defines the interface for recording and querying audit events.
type AuditRepository interface {
	// LogEvent records an audit event. Implementations should be append-only.
	LogEvent(ctx context.Context, event *entities.AuditEvent) error
	// Query retrieves audit events matching the given filter.
	Query(ctx context.Context, filter entities.AuditQueryFilter) ([]*entities.AuditEvent, int, error)
}
