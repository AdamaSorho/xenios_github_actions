package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryInsightCardRepository is an in-memory implementation of InsightCardRepository.
type InMemoryInsightCardRepository struct {
	mu       sync.RWMutex
	insights map[string]*entities.InsightCard
}

// NewInMemoryInsightCardRepository creates a new InMemoryInsightCardRepository.
func NewInMemoryInsightCardRepository() *InMemoryInsightCardRepository {
	return &InMemoryInsightCardRepository{
		insights: make(map[string]*entities.InsightCard),
	}
}

// Create stores a new insight card.
func (r *InMemoryInsightCardRepository) Create(_ context.Context, insight *entities.InsightCard) (*entities.InsightCard, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if insight.ID == "" {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			return nil, fmt.Errorf("generate id: %w", err)
		}
		insight.ID = hex.EncodeToString(b)
	}

	now := time.Now()
	insight.CreatedAt = now
	insight.UpdatedAt = now

	stored := *insight
	if stored.Evidence == nil {
		stored.Evidence = []entities.Evidence{}
	}
	r.insights[stored.ID] = &stored
	result := stored
	return &result, nil
}

// FindByID retrieves an insight card by its ID.
func (r *InMemoryInsightCardRepository) FindByID(_ context.Context, id string) (*entities.InsightCard, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	insight, ok := r.insights[id]
	if !ok {
		return nil, nil
	}
	result := *insight
	return &result, nil
}

// ListByCoachIDAndStatus retrieves insights for a coach filtered by status with pagination.
func (r *InMemoryInsightCardRepository) ListByCoachIDAndStatus(_ context.Context, coachID string, status entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []*entities.InsightCard
	for _, ic := range r.insights {
		if ic.CoachID == coachID && ic.Status == status {
			cp := *ic
			matched = append(matched, &cp)
		}
	}

	// Sort by priority rank then by created_at descending
	sort.Slice(matched, func(i, j int) bool {
		pi := entities.PriorityRank(matched[i].Priority)
		pj := entities.PriorityRank(matched[j].Priority)
		if pi != pj {
			return pi < pj
		}
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})

	total := len(matched)
	matched = applyPagination(matched, offset, limit)

	return matched, total, nil
}

// ListByClientID retrieves insights for a specific client with optional status filter.
func (r *InMemoryInsightCardRepository) ListByClientID(_ context.Context, clientID string, status *entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []*entities.InsightCard
	for _, ic := range r.insights {
		if ic.ClientID != clientID {
			continue
		}
		if status != nil && ic.Status != *status {
			continue
		}
		cp := *ic
		matched = append(matched, &cp)
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})

	total := len(matched)
	matched = applyPagination(matched, offset, limit)

	return matched, total, nil
}

// UpdateStatus updates the status and related timestamp of an insight card.
func (r *InMemoryInsightCardRepository) UpdateStatus(_ context.Context, id string, status entities.InsightStatus) (*entities.InsightCard, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ic, ok := r.insights[id]
	if !ok {
		return nil, nil
	}

	now := time.Now()
	ic.Status = status
	ic.UpdatedAt = now

	switch status {
	case entities.InsightStatusApproved:
		ic.ApprovedAt = &now
	case entities.InsightStatusDismissed:
		ic.DismissedAt = &now
	case entities.InsightStatusShared:
		ic.SharedAt = &now
	}

	result := *ic
	return &result, nil
}

// UpdateContent updates the title and body of an insight card.
func (r *InMemoryInsightCardRepository) UpdateContent(_ context.Context, id string, title, body string) (*entities.InsightCard, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ic, ok := r.insights[id]
	if !ok {
		return nil, nil
	}

	ic.Title = title
	ic.Body = body
	ic.UpdatedAt = time.Now()

	result := *ic
	return &result, nil
}

// CountByCoachIDAndStatus returns the count of insights for a coach with a given status.
func (r *InMemoryInsightCardRepository) CountByCoachIDAndStatus(_ context.Context, coachID string, status entities.InsightStatus) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, ic := range r.insights {
		if ic.CoachID == coachID && ic.Status == status {
			count++
		}
	}
	return count, nil
}

// applyPagination applies offset and limit to a slice of insight cards.
func applyPagination(items []*entities.InsightCard, offset, limit int) []*entities.InsightCard {
	if offset > len(items) {
		offset = len(items)
	}
	items = items[offset:]

	if limit <= 0 {
		limit = 20
	}
	if limit > len(items) {
		limit = len(items)
	}
	return items[:limit]
}
