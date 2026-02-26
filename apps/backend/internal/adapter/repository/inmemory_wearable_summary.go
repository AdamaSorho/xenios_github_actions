package repository

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryWearableSummaryRepository is an in-memory implementation of WearableSummaryRepository.
type InMemoryWearableSummaryRepository struct {
	mu      sync.RWMutex
	records []*entities.WearableSummary
}

// NewInMemoryWearableSummaryRepository creates a new in-memory wearable summary repository.
func NewInMemoryWearableSummaryRepository() *InMemoryWearableSummaryRepository {
	return &InMemoryWearableSummaryRepository{
		records: make([]*entities.WearableSummary, 0),
	}
}

// Upsert creates or updates a wearable summary for a given client/source/date.
func (r *InMemoryWearableSummaryRepository) Upsert(_ context.Context, summary *entities.WearableSummary) (*entities.WearableSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, rec := range r.records {
		if rec.ClientID == summary.ClientID && rec.Source == summary.Source && rec.SummaryDate == summary.SummaryDate {
			summary.ID = rec.ID
			summary.CreatedAt = rec.CreatedAt
			summary.SyncedAt = time.Now()
			r.records[i] = summary
			return summary, nil
		}
	}

	summary.ID = generateRepoID()
	summary.CreatedAt = time.Now()
	summary.SyncedAt = time.Now()
	r.records = append(r.records, summary)
	return summary, nil
}

// FindByClientID retrieves wearable summaries for a client, ordered by date descending.
func (r *InMemoryWearableSummaryRepository) FindByClientID(_ context.Context, clientID string, limit int) ([]*entities.WearableSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*entities.WearableSummary
	for _, rec := range r.records {
		if rec.ClientID == clientID {
			filtered = append(filtered, rec)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].SummaryDate > filtered[j].SummaryDate
	})

	if limit > 0 && limit < len(filtered) {
		filtered = filtered[:limit]
	}

	return filtered, nil
}

// Seed adds a wearable summary directly for testing purposes.
func (r *InMemoryWearableSummaryRepository) Seed(s *entities.WearableSummary) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, s)
}
