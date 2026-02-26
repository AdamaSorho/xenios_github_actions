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

// Upsert creates or updates a wearable summary for a given client, source, and date.
func (r *InMemoryWearableSummaryRepository) Upsert(_ context.Context, ws *entities.WearableSummary) (*entities.WearableSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, rec := range r.records {
		if rec.ClientID == ws.ClientID && rec.Source == ws.Source && rec.SummaryDate == ws.SummaryDate {
			updated := *ws
			updated.ID = rec.ID
			updated.CreatedAt = rec.CreatedAt
			updated.SyncedAt = time.Now()
			r.records[i] = &updated
			return &updated, nil
		}
	}

	created := *ws
	created.ID = generateRepoID()
	created.CreatedAt = time.Now()
	created.SyncedAt = time.Now()
	r.records = append(r.records, &created)
	return &created, nil
}

// FindByClientID retrieves wearable summaries for a client, ordered by date descending.
func (r *InMemoryWearableSummaryRepository) FindByClientID(_ context.Context, clientID string, limit, offset int) ([]*entities.WearableSummary, error) {
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

	if offset >= len(filtered) {
		return []*entities.WearableSummary{}, nil
	}

	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[offset:end], nil
}
