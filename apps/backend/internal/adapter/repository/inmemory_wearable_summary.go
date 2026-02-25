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
	summaries []*entities.WearableSummary
}

// NewInMemoryWearableSummaryRepository creates a new InMemoryWearableSummaryRepository.
func NewInMemoryWearableSummaryRepository() *InMemoryWearableSummaryRepository {
	return &InMemoryWearableSummaryRepository{}
}

// Add inserts a wearable summary into the in-memory store (for testing).
func (r *InMemoryWearableSummaryRepository) Add(s *entities.WearableSummary) {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := *s
	if stored.Metrics != nil {
		stored.Metrics = copyMetrics(s.Metrics)
	}
	r.summaries = append(r.summaries, &stored)
}

// FindByClientIDAndDateRange returns wearable summaries for a client within a date range.
func (r *InMemoryWearableSummaryRepository) FindByClientIDAndDateRange(_ context.Context, clientID string, from, to time.Time) ([]*entities.WearableSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entities.WearableSummary
	for _, s := range r.summaries {
		if s.ClientID == clientID &&
			!s.SummaryDate.Before(from) &&
			!s.SummaryDate.After(to) {
			cp := *s
			if s.Metrics != nil {
				cp.Metrics = copyMetrics(s.Metrics)
			}
			result = append(result, &cp)
		}
	}
	return result, nil
}

func copyMetrics(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{}, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}
