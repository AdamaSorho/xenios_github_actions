package repository

import (
	"context"
	"sync"
)

// AuditEvent is a recorded audit event (for testing/in-memory use).
type AuditEvent struct {
	ActorID    string
	Action     string
	EntityType string
	EntityID   string
	Metadata   map[string]interface{}
}

// InMemoryAuditRepository is an in-memory implementation of AuditRepository.
type InMemoryAuditRepository struct {
	mu     sync.RWMutex
	Events []AuditEvent
}

// NewInMemoryAuditRepository creates a new InMemoryAuditRepository.
func NewInMemoryAuditRepository() *InMemoryAuditRepository {
	return &InMemoryAuditRepository{}
}

// LogEvent records an audit event.
func (r *InMemoryAuditRepository) LogEvent(_ context.Context, actorID, action, entityType, entityID string, metadata map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Events = append(r.Events, AuditEvent{
		ActorID:    actorID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   metadata,
	})
	return nil
}
