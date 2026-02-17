package repository

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// InMemoryInsightCardRepository is an in-memory implementation of InsightCardRepository.
type InMemoryInsightCardRepository struct {
	mu      sync.RWMutex
	records []*entities.InsightCard
}

// NewInMemoryInsightCardRepository creates a new in-memory insight card repository.
func NewInMemoryInsightCardRepository() *InMemoryInsightCardRepository {
	return &InMemoryInsightCardRepository{
		records: make([]*entities.InsightCard, 0),
	}
}

// Create stores a new insight card.
func (r *InMemoryInsightCardRepository) Create(_ context.Context, insight *entities.InsightCard) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if insight.ID == "" {
		insight.ID = generateRepoID()
	}
	// Store a copy
	cp := *insight
	r.records = append(r.records, &cp)
	return nil
}

// FindByID retrieves a single insight card by its ID.
func (r *InMemoryInsightCardRepository) FindByID(_ context.Context, id string) (*entities.InsightCard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rec := range r.records {
		if rec.ID == id {
			cp := *rec
			return &cp, nil
		}
	}
	return nil, nil
}

// Update persists changes to an existing insight card.
func (r *InMemoryInsightCardRepository) Update(_ context.Context, insight *entities.InsightCard) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, rec := range r.records {
		if rec.ID == insight.ID {
			cp := *insight
			r.records[i] = &cp
			return nil
		}
	}
	return fmt.Errorf("insight card %s not found", insight.ID)
}

// ListByCoach retrieves insight cards for a coach, sorted by priority (desc) then creation time (desc).
func (r *InMemoryInsightCardRepository) ListByCoach(_ context.Context, filter repository.InsightCardFilter) ([]*entities.InsightCard, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*entities.InsightCard
	for _, rec := range r.records {
		if rec.CoachID != filter.CoachID {
			continue
		}
		if filter.Status != "" && rec.Status != filter.Status {
			continue
		}
		if filter.ClientID != "" && rec.ClientID != filter.ClientID {
			continue
		}
		cp := *rec
		filtered = append(filtered, &cp)
	}

	// Sort by priority (desc), then by creation time (desc)
	sort.Slice(filtered, func(i, j int) bool {
		ri := entities.InsightPriorityRank(filtered[i].Priority)
		rj := entities.InsightPriorityRank(filtered[j].Priority)
		if ri != rj {
			return ri > rj
		}
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	total := len(filtered)
	if filter.Offset >= total {
		return []*entities.InsightCard{}, total, nil
	}
	end := filter.Offset + filter.Limit
	if end > total {
		end = total
	}
	return filtered[filter.Offset:end], total, nil
}

// ListByClient retrieves insight cards for a specific client.
func (r *InMemoryInsightCardRepository) ListByClient(_ context.Context, filter repository.InsightCardFilter) ([]*entities.InsightCard, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*entities.InsightCard
	for _, rec := range r.records {
		if rec.ClientID != filter.ClientID {
			continue
		}
		if filter.CoachID != "" && rec.CoachID != filter.CoachID {
			continue
		}
		if filter.Status != "" && rec.Status != filter.Status {
			continue
		}
		cp := *rec
		filtered = append(filtered, &cp)
	}

	// Sort by creation time (desc)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	total := len(filtered)
	if filter.Offset >= total {
		return []*entities.InsightCard{}, total, nil
	}
	end := filter.Offset + filter.Limit
	if end > total {
		end = total
	}
	return filtered[filter.Offset:end], total, nil
}
