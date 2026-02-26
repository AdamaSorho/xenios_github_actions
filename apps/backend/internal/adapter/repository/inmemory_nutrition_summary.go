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

// InMemoryNutritionSummaryRepository is an in-memory implementation of NutritionSummaryRepository.
type InMemoryNutritionSummaryRepository struct {
	mu        sync.RWMutex
	summaries map[string]*entities.NutritionSummary // keyed by clientID+artifactID
}

// NewInMemoryNutritionSummaryRepository creates a new InMemoryNutritionSummaryRepository.
func NewInMemoryNutritionSummaryRepository() *InMemoryNutritionSummaryRepository {
	return &InMemoryNutritionSummaryRepository{
		summaries: make(map[string]*entities.NutritionSummary),
	}
}

// Upsert creates or updates a nutrition summary.
func (r *InMemoryNutritionSummaryRepository) Upsert(_ context.Context, summary *entities.NutritionSummary) (*entities.NutritionSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := summary.ClientID + ":" + summary.ArtifactID
	now := time.Now()

	existing, ok := r.summaries[key]
	if ok {
		summary.ID = existing.ID
		summary.CreatedAt = existing.CreatedAt
	} else {
		if summary.ID == "" {
			b := make([]byte, 16)
			if _, err := rand.Read(b); err != nil {
				return nil, fmt.Errorf("generate id: %w", err)
			}
			summary.ID = hex.EncodeToString(b)
		}
		summary.CreatedAt = now
	}
	summary.UpdatedAt = now

	stored := *summary
	r.summaries[key] = &stored
	result := stored
	return &result, nil
}

// GetAll returns all stored summaries (for testing).
func (r *InMemoryNutritionSummaryRepository) GetAll() []*entities.NutritionSummary {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*entities.NutritionSummary, 0, len(r.summaries))
	for _, s := range r.summaries {
		val := *s
		result = append(result, &val)
	}
	return result
}
