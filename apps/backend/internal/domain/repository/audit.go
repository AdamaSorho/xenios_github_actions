package repository

import "context"

// AuditRepository defines the interface for recording audit events.
type AuditRepository interface {
	LogEvent(ctx context.Context, actorID, action, entityType, entityID string, metadata map[string]interface{}) error
}
