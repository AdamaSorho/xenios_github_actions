package repository

import (
	"context"
	"sync"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryWearableSummaryRepository is a thread-safe in-memory implementation
// of WearableSummaryRepository for testing.
type InMemoryWearableSummaryRepository struct {
	mu        sync.RWMutex
	summaries map[string]*entities.WearableSummary // key: clientID+source+date
}

// NewInMemoryWearableSummaryRepository creates a new in-memory wearable summary repository.
func NewInMemoryWearableSummaryRepository() *InMemoryWearableSummaryRepository {
	return &InMemoryWearableSummaryRepository{
		summaries: make(map[string]*entities.WearableSummary),
	}
}

// Upsert inserts or updates a wearable summary.
func (r *InMemoryWearableSummaryRepository) Upsert(_ context.Context, summary *entities.WearableSummary) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := summary.ClientID + "|" + string(summary.Source) + "|" + summary.SummaryDate.Format("2006-01-02")
	r.summaries[key] = summary
	return nil
}

// GetSummaries returns all stored summaries (for test assertions).
func (r *InMemoryWearableSummaryRepository) GetSummaries() []*entities.WearableSummary {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*entities.WearableSummary, 0, len(r.summaries))
	for _, s := range r.summaries {
		out = append(out, s)
	}
	return out
}
