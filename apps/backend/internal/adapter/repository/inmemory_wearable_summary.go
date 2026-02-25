package repository

import (
	"context"
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

// Upsert creates or updates a wearable summary.
func (r *InMemoryWearableSummaryRepository) Upsert(_ context.Context, summary *entities.WearableSummary) (*entities.WearableSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for existing entry with same client/source/date
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

// FindByClientID retrieves wearable summaries for a client within the last N days.
func (r *InMemoryWearableSummaryRepository) FindByClientID(_ context.Context, clientID string, days int) ([]*entities.WearableSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cutoff := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	var result []*entities.WearableSummary
	for _, rec := range r.records {
		if rec.ClientID == clientID && rec.SummaryDate >= cutoff {
			result = append(result, rec)
		}
	}
	return result, nil
}
