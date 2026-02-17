package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryInsightCardRepository is an in-memory implementation of InsightCardRepository.
type InMemoryInsightCardRepository struct {
	mu    sync.RWMutex
	cards []*entities.InsightCard
	count int
}

// NewInMemoryInsightCardRepository creates a new InMemoryInsightCardRepository.
func NewInMemoryInsightCardRepository() *InMemoryInsightCardRepository {
	return &InMemoryInsightCardRepository{}
}

// Create stores a new insight card and returns it with generated ID and timestamps.
func (r *InMemoryInsightCardRepository) Create(_ context.Context, card *entities.InsightCard) (*entities.InsightCard, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.count++
	stored := *card
	if stored.ID == "" {
		stored.ID = fmt.Sprintf("insight-%d", r.count)
	}
	now := time.Now()
	stored.CreatedAt = now
	stored.UpdatedAt = now

	// Deep copy evidence slice
	if card.Evidence != nil {
		stored.Evidence = make([]entities.EvidenceRef, len(card.Evidence))
		copy(stored.Evidence, card.Evidence)
	}

	r.cards = append(r.cards, &stored)
	return &stored, nil
}

// FindByClientID returns insight cards for a given client with pagination.
func (r *InMemoryInsightCardRepository) FindByClientID(_ context.Context, clientID string, limit, offset int) ([]*entities.InsightCard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []*entities.InsightCard
	for _, c := range r.cards {
		if c.ClientID == clientID {
			matched = append(matched, c)
		}
	}

	if offset > len(matched) {
		offset = len(matched)
	}
	matched = matched[offset:]

	if limit <= 0 {
		limit = 50
	}
	if limit > len(matched) {
		limit = len(matched)
	}
	return matched[:limit], nil
}

// FindByStatus returns insight cards matching a given status with pagination.
func (r *InMemoryInsightCardRepository) FindByStatus(_ context.Context, status entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []*entities.InsightCard
	for _, c := range r.cards {
		if c.Status == status {
			matched = append(matched, c)
		}
	}

	if offset > len(matched) {
		offset = len(matched)
	}
	matched = matched[offset:]

	if limit <= 0 {
		limit = 50
	}
	if limit > len(matched) {
		limit = len(matched)
	}
	return matched[:limit], nil
}

// UpdateStatus changes the status of an insight card by ID.
func (r *InMemoryInsightCardRepository) UpdateStatus(_ context.Context, id string, status entities.InsightStatus) (*entities.InsightCard, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, c := range r.cards {
		if c.ID == id {
			c.Status = status
			c.UpdatedAt = time.Now()
			return c, nil
		}
	}
	return nil, fmt.Errorf("insight card not found: %s", id)
}

// ExistsByMeasurementID returns true if an insight card already exists
// that references the given measurement ID in its evidence.
func (r *InMemoryInsightCardRepository) ExistsByMeasurementID(_ context.Context, measurementID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.cards {
		for _, ev := range c.Evidence {
			if ev.MeasurementID == measurementID {
				return true, nil
			}
		}
	}
	return false, nil
}

// CardCount returns the number of stored cards (for testing).
func (r *InMemoryInsightCardRepository) CardCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.cards)
}

// GetCards returns a snapshot copy of all stored cards (for testing).
func (r *InMemoryInsightCardRepository) GetCards() []*entities.InsightCard {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make([]*entities.InsightCard, len(r.cards))
	copy(cp, r.cards)
	return cp
}
