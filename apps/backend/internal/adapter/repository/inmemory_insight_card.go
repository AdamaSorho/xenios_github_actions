package repository

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryInsightCardRepository is an in-memory implementation of InsightCardRepository.
type InMemoryInsightCardRepository struct {
	mu    sync.RWMutex
	cards map[string]*entities.InsightCard
}

// NewInMemoryInsightCardRepository creates a new in-memory insight card repository.
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
		card.ID = generateRepoID()
	}
	now := time.Now()
	card.CreatedAt = now
	card.UpdatedAt = now

	stored := *card
	if card.Evidence != nil {
		stored.Evidence = make([]entities.InsightEvidence, len(card.Evidence))
		copy(stored.Evidence, card.Evidence)
	}
	r.cards[stored.ID] = &stored

	result := stored
	return &result, nil
}

// FindByID retrieves an insight card by its ID.
func (r *InMemoryInsightCardRepository) FindByID(_ context.Context, id string) (*entities.InsightCard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	card, ok := r.cards[id]
	if !ok {
		return nil, nil
	}
	result := *card
	return &result, nil
}

// Update persists changes to an existing insight card.
func (r *InMemoryInsightCardRepository) Update(_ context.Context, card *entities.InsightCard) (*entities.InsightCard, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.cards[card.ID]; !ok {
		return nil, nil
	}

	card.UpdatedAt = time.Now()
	stored := *card
	if card.Evidence != nil {
		stored.Evidence = make([]entities.InsightEvidence, len(card.Evidence))
		copy(stored.Evidence, card.Evidence)
	}
	r.cards[stored.ID] = &stored

	result := stored
	return &result, nil
}

// ListByCoachID retrieves insight cards for a coach, filtered by optional status.
func (r *InMemoryInsightCardRepository) ListByCoachID(_ context.Context, filter entities.InsightQueryFilter) ([]*entities.InsightCard, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*entities.InsightCard
	for _, card := range r.cards {
		if card.CoachID != filter.CoachID {
			continue
		}
		if filter.Status != "" && string(card.Status) != filter.Status {
			continue
		}
		c := *card
		filtered = append(filtered, &c)
	}

	sortInsights(filtered)
	total := len(filtered)
	filtered = paginate(filtered, filter.Page, filter.Limit)

	return filtered, total, nil
}

// ListByClientID retrieves insight cards for a client, filtered by optional status.
func (r *InMemoryInsightCardRepository) ListByClientID(_ context.Context, filter entities.InsightQueryFilter) ([]*entities.InsightCard, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*entities.InsightCard
	for _, card := range r.cards {
		if card.ClientID != filter.ClientID {
			continue
		}
		if filter.Status != "" && string(card.Status) != filter.Status {
			continue
		}
		c := *card
		filtered = append(filtered, &c)
	}

	sortInsights(filtered)
	total := len(filtered)
	filtered = paginate(filtered, filter.Page, filter.Limit)

	return filtered, total, nil
}

func sortInsights(cards []*entities.InsightCard) {
	sort.Slice(cards, func(i, j int) bool {
		pi := cards[i].Priority.SortOrder()
		pj := cards[j].Priority.SortOrder()
		if pi != pj {
			return pi < pj
		}
		return cards[i].CreatedAt.After(cards[j].CreatedAt)
	})
}

func paginate(cards []*entities.InsightCard, page, limit int) []*entities.InsightCard {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	if offset >= len(cards) {
		return []*entities.InsightCard{}
	}
	end := offset + limit
	if end > len(cards) {
		end = len(cards)
	}
	return cards[offset:end]
}
