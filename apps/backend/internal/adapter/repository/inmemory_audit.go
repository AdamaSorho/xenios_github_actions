package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryAuditRepository is an in-memory implementation of AuditRepository.
type InMemoryAuditRepository struct {
	mu     sync.RWMutex
	Events []*entities.AuditEvent
	count  int
}

// NewInMemoryAuditRepository creates a new InMemoryAuditRepository.
func NewInMemoryAuditRepository() *InMemoryAuditRepository {
	return &InMemoryAuditRepository{}
}

// LogEvent records an audit event.
func (r *InMemoryAuditRepository) LogEvent(_ context.Context, event *entities.AuditEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.count++
	stored := *event
	if stored.ID == "" {
		stored.ID = fmt.Sprintf("audit-%d", r.count)
	}
	if stored.CreatedAt.IsZero() {
		stored.CreatedAt = time.Now()
	}

	r.Events = append(r.Events, &stored)
	return nil
}

// Query returns audit events matching the filter.
func (r *InMemoryAuditRepository) Query(_ context.Context, filter entities.AuditQueryFilter) ([]*entities.AuditEvent, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []*entities.AuditEvent
	for _, e := range r.Events {
		if filter.ActorID != "" && e.ActorID != filter.ActorID {
			continue
		}
		if filter.Action != "" && e.Action != filter.Action {
			continue
		}
		if filter.EntityType != "" && e.EntityType != filter.EntityType {
			continue
		}
		if filter.EntityID != "" && e.EntityID != filter.EntityID {
			continue
		}
		if filter.From != nil && e.CreatedAt.Before(*filter.From) {
			continue
		}
		if filter.To != nil && e.CreatedAt.After(*filter.To) {
			continue
		}
		matched = append(matched, e)
	}

	total := len(matched)

	// Apply pagination
	offset := filter.Offset
	if offset > len(matched) {
		offset = len(matched)
	}
	matched = matched[offset:]

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > len(matched) {
		limit = len(matched)
	}
	matched = matched[:limit]

	return matched, total, nil
}
