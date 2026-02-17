package repository

import (
	"context"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryWearableSummaryRepository is an in-memory implementation of WearableSummaryRepository.
type InMemoryWearableSummaryRepository struct {
	mu        sync.RWMutex
	summaries map[string]*entities.WearableSummary // keyed by clientID+source+date
}

// NewInMemoryWearableSummaryRepository creates a new InMemoryWearableSummaryRepository.
func NewInMemoryWearableSummaryRepository() *InMemoryWearableSummaryRepository {
	return &InMemoryWearableSummaryRepository{
		summaries: make(map[string]*entities.WearableSummary),
	}
}

// Upsert creates or updates a wearable summary.
func (r *InMemoryWearableSummaryRepository) Upsert(_ context.Context, summary *entities.WearableSummary) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := summaryKey(summary.ClientID, summary.Source, summary.SummaryDate)

	if existing, ok := r.summaries[key]; ok {
		existing.Metrics = summary.Metrics
		existing.SyncedAt = time.Now()
	} else {
		stored := *summary
		if stored.CreatedAt.IsZero() {
			stored.CreatedAt = time.Now()
		}
		stored.SyncedAt = time.Now()
		r.summaries[key] = &stored
	}
	return nil
}

// GetAll returns all stored summaries (for testing).
func (r *InMemoryWearableSummaryRepository) GetAll() []*entities.WearableSummary {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*entities.WearableSummary, 0, len(r.summaries))
	for _, s := range r.summaries {
		cp := *s
		result = append(result, &cp)
	}
	return result
}

// Count returns the number of stored summaries (for testing).
func (r *InMemoryWearableSummaryRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.summaries)
}

func summaryKey(clientID string, source entities.WearableSource, date time.Time) string {
	return clientID + "|" + string(source) + "|" + date.Format("2006-01-02")
}
