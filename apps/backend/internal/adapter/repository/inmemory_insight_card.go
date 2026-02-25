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
	mu    sync.RWMutex
	cards []*entities.InsightCard
}

// NewInMemoryInsightCardRepository creates a new in-memory insight card repository.
func NewInMemoryInsightCardRepository() *InMemoryInsightCardRepository {
	return &InMemoryInsightCardRepository{
		cards: make([]*entities.InsightCard, 0),
	}
}

// FindByID retrieves a single insight card by ID.
func (r *InMemoryInsightCardRepository) FindByID(_ context.Context, id string) (*entities.InsightCard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.cards {
		if c.ID == id {
			copy := *c
			return &copy, nil
		}
	}
	return nil, nil
}

// Create stores a new insight card.
func (r *InMemoryInsightCardRepository) Create(_ context.Context, card *entities.InsightCard) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if card.ID == "" {
		card.ID = generateRepoID()
	}
	copy := *card
	r.cards = append(r.cards, &copy)
	return nil
}

// Update persists changes to an existing insight card.
func (r *InMemoryInsightCardRepository) Update(_ context.Context, card *entities.InsightCard) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, c := range r.cards {
		if c.ID == card.ID {
			copy := *card
			r.cards[i] = &copy
			return nil
		}
	}
	return fmt.Errorf("insight card not found: %s", card.ID)
}

// priorityOrder maps priority strings to sort order (lower = higher priority).
var priorityOrder = map[string]int{
	entities.InsightPriorityUrgent: 0,
	entities.InsightPriorityHigh:   1,
	entities.InsightPriorityMedium: 2,
	entities.InsightPriorityLow:    3,
}

// ListByCoach retrieves insight cards for a coach with filtering and pagination.
func (r *InMemoryInsightCardRepository) ListByCoach(_ context.Context, filter repository.InsightCardFilter) ([]*entities.InsightCard, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*entities.InsightCard
	for _, c := range r.cards {
		if c.CoachID != filter.CoachID {
			continue
		}
		if filter.Status != "" && c.Status != filter.Status {
			continue
		}
		if filter.ClientID != "" && c.ClientID != filter.ClientID {
			continue
		}
		copy := *c
		filtered = append(filtered, &copy)
	}

	// Sort by priority (urgent first), then by created_at (newest first)
	sort.Slice(filtered, func(i, j int) bool {
		pi := priorityOrder[filtered[i].Priority]
		pj := priorityOrder[filtered[j].Priority]
		if pi != pj {
			return pi < pj
		}
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	total := len(filtered)

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	if offset >= len(filtered) {
		return []*entities.InsightCard{}, total, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[offset:end], total, nil
}

// ListByClient retrieves insight cards for a specific client with filtering.
func (r *InMemoryInsightCardRepository) ListByClient(_ context.Context, filter repository.InsightCardFilter) ([]*entities.InsightCard, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*entities.InsightCard
	for _, c := range r.cards {
		if c.ClientID != filter.ClientID {
			continue
		}
		if filter.CoachID != "" && c.CoachID != filter.CoachID {
			continue
		}
		if filter.Status != "" && c.Status != filter.Status {
			continue
		}
		copy := *c
		filtered = append(filtered, &copy)
	}

	// Sort by created_at (newest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	total := len(filtered)

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	if offset >= len(filtered) {
		return []*entities.InsightCard{}, total, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[offset:end], total, nil
}

// Seed adds insight cards directly for testing purposes.
func (r *InMemoryInsightCardRepository) Seed(cards ...*entities.InsightCard) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range cards {
		copy := *c
		if copy.ID == "" {
			copy.ID = generateRepoID()
		}
		r.cards = append(r.cards, &copy)
	}
}
