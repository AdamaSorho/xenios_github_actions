package repository

import (
	"context"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryCoachClientRepository is an in-memory implementation of CoachClientRepository.
// This is a skeleton placeholder — a PostgreSQL implementation will follow with the migration PR.
type InMemoryCoachClientRepository struct {
	mu      sync.RWMutex
	records []*entities.CoachClient
}

// NewInMemoryCoachClientRepository creates a new in-memory coach-client repository.
func NewInMemoryCoachClientRepository() *InMemoryCoachClientRepository {
	return &InMemoryCoachClientRepository{
		records: make([]*entities.CoachClient, 0),
	}
}

// Create stores a new coach-client relationship.
func (r *InMemoryCoachClientRepository) Create(_ context.Context, coachID, clientID string) (*entities.CoachClient, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id, _ := generateID()
	cc := &entities.CoachClient{
		ID:        id,
		CoachID:   coachID,
		ClientID:  clientID,
		CreatedAt: time.Now(),
	}
	r.records = append(r.records, cc)
	return cc, nil
}

// ListByCoachID retrieves all clients for a given coach with pagination.
func (r *InMemoryCoachClientRepository) ListByCoachID(_ context.Context, coachID string, limit, offset int) ([]*entities.CoachClient, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*entities.CoachClient
	for _, rec := range r.records {
		if rec.CoachID == coachID {
			filtered = append(filtered, rec)
		}
	}

	if offset >= len(filtered) {
		return []*entities.CoachClient{}, nil
	}

	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[offset:end], nil
}

