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

// Create stores a new insight card and returns it with a generated ID.
func (r *InMemoryInsightCardRepository) Create(_ context.Context, card *entities.InsightCard) (*entities.InsightCard, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.count++
	stored := *card
	if stored.ID == "" {
		stored.ID = fmt.Sprintf("insight-%d", r.count)
	}
	now := time.Now()
	if stored.CreatedAt.IsZero() {
		stored.CreatedAt = now
	}
	if stored.UpdatedAt.IsZero() {
		stored.UpdatedAt = now
	}

	evidence := make([]entities.EvidenceRef, len(card.Evidence))
	copy(evidence, card.Evidence)
	stored.Evidence = evidence

	r.cards = append(r.cards, &stored)
	return &stored, nil
}

// FindByID returns the insight card with the given ID, or nil if not found.
func (r *InMemoryInsightCardRepository) FindByID(_ context.Context, id string) (*entities.InsightCard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.cards {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, nil
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

	return paginate(matched, limit, offset), nil
}

// FindByStatus returns insight cards filtered by coach and status with pagination.
func (r *InMemoryInsightCardRepository) FindByStatus(_ context.Context, coachID string, status entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []*entities.InsightCard
	for _, c := range r.cards {
		if c.CoachID == coachID && c.Status == status {
			matched = append(matched, c)
		}
	}

	return paginate(matched, limit, offset), nil
}

// ExistsByEvidence checks whether an insight already exists for a given client and measurement.
func (r *InMemoryInsightCardRepository) ExistsByEvidence(_ context.Context, clientID string, measurementID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.cards {
		if c.ClientID != clientID {
			continue
		}
		for _, ev := range c.Evidence {
			if ev.MeasurementID == measurementID {
				return true, nil
			}
		}
	}
	return false, nil
}

// UpdateStatus updates the status of an insight card.
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

// CardCount returns the number of stored cards (for testing).
func (r *InMemoryInsightCardRepository) CardCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.cards)
}

// GetCards returns a snapshot of all stored cards (for testing).
func (r *InMemoryInsightCardRepository) GetCards() []*entities.InsightCard {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make([]*entities.InsightCard, len(r.cards))
	copy(cp, r.cards)
	return cp
}

func paginate(items []*entities.InsightCard, limit, offset int) []*entities.InsightCard {
	if offset > len(items) {
		offset = len(items)
	}
	items = items[offset:]

	if limit <= 0 {
		limit = 50
	}
	if limit > len(items) {
		limit = len(items)
	}
	return items[:limit]
}
