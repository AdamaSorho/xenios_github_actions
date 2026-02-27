package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryNutritionSummaryRepository is an in-memory implementation of NutritionSummaryRepository for testing.
type InMemoryNutritionSummaryRepository struct {
	mu        sync.RWMutex
	summaries map[string]*entities.NutritionSummary
	count     int
}

// NewInMemoryNutritionSummaryRepository creates a new InMemoryNutritionSummaryRepository.
func NewInMemoryNutritionSummaryRepository() *InMemoryNutritionSummaryRepository {
	return &InMemoryNutritionSummaryRepository{
		summaries: make(map[string]*entities.NutritionSummary),
	}
}

// Upsert creates or updates a nutrition summary keyed by client_id + artifact_id.
func (r *InMemoryNutritionSummaryRepository) Upsert(_ context.Context, summary *entities.NutritionSummary) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := summary.ClientID + ":" + summary.ArtifactID
	now := time.Now()

	existing, ok := r.summaries[key]
	if ok {
		summary.ID = existing.ID
		summary.CreatedAt = existing.CreatedAt
		summary.UpdatedAt = now
	} else {
		r.count++
		summary.ID = fmt.Sprintf("ns-%d", r.count)
		summary.CreatedAt = now
		summary.UpdatedAt = now
	}

	stored := *summary
	r.summaries[key] = &stored
	return nil
}

// GetByClientID returns all summaries for a client (for testing).
func (r *InMemoryNutritionSummaryRepository) GetByClientID(clientID string) []*entities.NutritionSummary {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entities.NutritionSummary
	for _, s := range r.summaries {
		if s.ClientID == clientID {
			cp := *s
			result = append(result, &cp)
		}
	}
	return result
}
