package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"

	"github.com/xenios/backend/internal/domain"
)

// InMemoryCoachClientRepository is a temporary in-memory implementation of
// CoachClientRepository for the API skeleton. It will be replaced by
// PostgresCoachClientRepository when the database migration is implemented.
type InMemoryCoachClientRepository struct {
	mu      sync.RWMutex
	records []*domain.CoachClient
}

// NewInMemoryCoachClientRepository creates a new in-memory coach-client repository.
func NewInMemoryCoachClientRepository() *InMemoryCoachClientRepository {
	return &InMemoryCoachClientRepository{
		records: make([]*domain.CoachClient, 0),
	}
}

// Create adds a new coach-client relationship.
func (r *InMemoryCoachClientRepository) Create(_ context.Context, cc *domain.CoachClient) (*domain.CoachClient, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cc.ID = generateRepoID()
	r.records = append(r.records, cc)
	return cc, nil
}

// ListByCoachID returns clients for a given coach with pagination.
func (r *InMemoryCoachClientRepository) ListByCoachID(_ context.Context, coachID string, limit, offset int) ([]*domain.CoachClient, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*domain.CoachClient
	for _, rec := range r.records {
		if rec.CoachID == coachID {
			filtered = append(filtered, rec)
		}
	}

	// Apply pagination
	if offset >= len(filtered) {
		return []*domain.CoachClient{}, nil
	}

	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[offset:end], nil
}

func generateRepoID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}
