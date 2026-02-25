package repository

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryWearableSummaryRepository is a test-friendly in-memory implementation.
type InMemoryWearableSummaryRepository struct {
	mu        sync.RWMutex
	summaries map[string]json.RawMessage // key: "clientID:source"
}

// NewInMemoryWearableSummaryRepository creates a new in-memory wearable summary store.
func NewInMemoryWearableSummaryRepository() *InMemoryWearableSummaryRepository {
	return &InMemoryWearableSummaryRepository{
		summaries: make(map[string]json.RawMessage),
	}
}

func (r *InMemoryWearableSummaryRepository) Upsert(_ context.Context, clientID string, source entities.WearableSource, metrics json.RawMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := clientID + ":" + string(source)
	r.summaries[key] = metrics
	return nil
}

// Get returns the stored metrics for testing purposes.
func (r *InMemoryWearableSummaryRepository) Get(clientID string, source entities.WearableSource) (json.RawMessage, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := clientID + ":" + string(source)
	m, ok := r.summaries[key]
	return m, ok
}
