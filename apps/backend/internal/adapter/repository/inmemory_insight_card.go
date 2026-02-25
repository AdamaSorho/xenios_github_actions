package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryInsightCardRepository is an in-memory implementation of InsightCardRepository.
type InMemoryInsightCardRepository struct {
	mu    sync.RWMutex
	cards map[string]*entities.InsightCard
}

// NewInMemoryInsightCardRepository creates a new InMemoryInsightCardRepository.
func NewInMemoryInsightCardRepository() *InMemoryInsightCardRepository {
	return &InMemoryInsightCardRepository{
		cards: make(map[string]*entities.InsightCard),
	}
}

// Create stores a new insight card.
func (r *InMemoryInsightCardRepository) Create(_ context.Context, card *entities.InsightCard) (*entities.InsightCard, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if card.ID == "" {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			return nil, fmt.Errorf("generate id: %w", err)
		}
		card.ID = hex.EncodeToString(b)
	}

	now := time.Now()
	card.CreatedAt = now
	card.UpdatedAt = now

	stored := *card
	stored.Evidence = make([]entities.EvidenceRef, len(card.Evidence))
	copy(stored.Evidence, card.Evidence)

	r.cards[stored.ID] = &stored

	result := stored
	result.Evidence = make([]entities.EvidenceRef, len(stored.Evidence))
	copy(result.Evidence, stored.Evidence)
	return &result, nil
}

// FindByClientID returns all insight cards for a given client.
func (r *InMemoryInsightCardRepository) FindByClientID(_ context.Context, clientID string) ([]*entities.InsightCard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entities.InsightCard
	for _, card := range r.cards {
		if card.ClientID == clientID {
			cp := *card
			cp.Evidence = make([]entities.EvidenceRef, len(card.Evidence))
			copy(cp.Evidence, card.Evidence)
			result = append(result, &cp)
		}
	}
	return result, nil
}

// FindByStatus returns all insight cards with the given status.
func (r *InMemoryInsightCardRepository) FindByStatus(_ context.Context, status entities.InsightStatus) ([]*entities.InsightCard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entities.InsightCard
	for _, card := range r.cards {
		if card.Status == status {
			cp := *card
			cp.Evidence = make([]entities.EvidenceRef, len(card.Evidence))
			copy(cp.Evidence, card.Evidence)
			result = append(result, &cp)
		}
	}
	return result, nil
}

// UpdateStatus updates the status of an insight card.
func (r *InMemoryInsightCardRepository) UpdateStatus(_ context.Context, id string, status entities.InsightStatus) (*entities.InsightCard, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	card, ok := r.cards[id]
	if !ok {
		return nil, fmt.Errorf("insight card not found: %s", id)
	}

	card.Status = status
	card.UpdatedAt = time.Now()

	result := *card
	result.Evidence = make([]entities.EvidenceRef, len(card.Evidence))
	copy(result.Evidence, card.Evidence)
	return &result, nil
}

// ExistsByEvidence checks if an insight card already exists referencing the given measurement.
func (r *InMemoryInsightCardRepository) ExistsByEvidence(_ context.Context, clientID string, measurementID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, card := range r.cards {
		if card.ClientID != clientID {
			continue
		}
		for _, ev := range card.Evidence {
			if ev.MeasurementID == measurementID {
				return true, nil
			}
		}
	}
	return false, nil
}

// GetAll returns all stored insight cards (for testing).
func (r *InMemoryInsightCardRepository) GetAll() []*entities.InsightCard {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*entities.InsightCard, 0, len(r.cards))
	for _, card := range r.cards {
		cp := *card
		cp.Evidence = make([]entities.EvidenceRef, len(card.Evidence))
		copy(cp.Evidence, card.Evidence)
		result = append(result, &cp)
	}
	return result
}
