package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryNutritionRepository is an in-memory implementation of NutritionRepository.
type InMemoryNutritionRepository struct {
	mu        sync.RWMutex
	records   []*entities.NutritionRecord
	summaries map[string]*entities.NutritionSummary // keyed by clientID+artifactID
	count     int
}

// NewInMemoryNutritionRepository creates a new InMemoryNutritionRepository.
func NewInMemoryNutritionRepository() *InMemoryNutritionRepository {
	return &InMemoryNutritionRepository{
		summaries: make(map[string]*entities.NutritionSummary),
	}
}

// SaveRecords stores a batch of daily nutrition records.
func (r *InMemoryNutritionRepository) SaveRecords(_ context.Context, records []*entities.NutritionRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, rec := range records {
		r.count++
		stored := *rec
		if stored.ID == "" {
			stored.ID = fmt.Sprintf("nutrition-rec-%d", r.count)
		}
		if stored.CreatedAt.IsZero() {
			stored.CreatedAt = time.Now()
		}
		r.records = append(r.records, &stored)
	}
	return nil
}

// UpsertSummary creates or updates a nutrition summary.
func (r *InMemoryNutritionRepository) UpsertSummary(_ context.Context, summary *entities.NutritionSummary) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := summary.ClientID + ":" + summary.ArtifactID
	stored := *summary
	if stored.ID == "" {
		stored.ID = fmt.Sprintf("nutrition-sum-%d", len(r.summaries)+1)
	}
	now := time.Now()
	if stored.CreatedAt.IsZero() {
		stored.CreatedAt = now
	}
	stored.UpdatedAt = now
	if stored.ComputedAt.IsZero() {
		stored.ComputedAt = now
	}
	r.summaries[key] = &stored
	return nil
}

// GetSummaryByClientID retrieves the latest nutrition summary for a client.
func (r *InMemoryNutritionRepository) GetSummaryByClientID(_ context.Context, clientID string) (*entities.NutritionSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latest *entities.NutritionSummary
	for _, s := range r.summaries {
		if s.ClientID == clientID {
			if latest == nil || s.ComputedAt.After(latest.ComputedAt) {
				latest = s
			}
		}
	}
	if latest == nil {
		return nil, nil
	}
	result := *latest
	return &result, nil
}

// GetRecords returns all stored records (for testing).
func (r *InMemoryNutritionRepository) GetRecords() []*entities.NutritionRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make([]*entities.NutritionRecord, len(r.records))
	copy(cp, r.records)
	return cp
}

// GetSummaries returns all stored summaries (for testing).
func (r *InMemoryNutritionRepository) GetSummaries() []*entities.NutritionSummary {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*entities.NutritionSummary, 0, len(r.summaries))
	for _, s := range r.summaries {
		cp := *s
		result = append(result, &cp)
	}
	return result
}
