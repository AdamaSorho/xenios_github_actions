package repository

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryMeasurementRepository is an in-memory implementation of MeasurementRepository.
type InMemoryMeasurementRepository struct {
	mu      sync.RWMutex
	records []*entities.Measurement
}

// NewInMemoryMeasurementRepository creates a new in-memory measurement repository.
func NewInMemoryMeasurementRepository() *InMemoryMeasurementRepository {
	return &InMemoryMeasurementRepository{
		records: make([]*entities.Measurement, 0),
	}
}

// Create stores a new measurement.
func (r *InMemoryMeasurementRepository) Create(_ context.Context, m *entities.Measurement) (*entities.Measurement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	created := *m
	created.ID = generateRepoID()
	created.CreatedAt = time.Now()
	r.records = append(r.records, &created)
	return &created, nil
}

// FindByClientID retrieves measurements for a client with filtering and pagination.
func (r *InMemoryMeasurementRepository) FindByClientID(_ context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*entities.Measurement
	for _, rec := range r.records {
		if rec.ClientID != filter.ClientID {
			continue
		}
		if filter.MeasurementType != "" && rec.MeasurementType != filter.MeasurementType {
			continue
		}
		if filter.From != nil && rec.MeasuredAt.Before(*filter.From) {
			continue
		}
		if filter.To != nil && rec.MeasuredAt.After(*filter.To) {
			continue
		}
		filtered = append(filtered, rec)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].MeasuredAt.After(filtered[j].MeasuredAt)
	})

	total := len(filtered)
	if filter.Offset >= total {
		return []*entities.Measurement{}, total, nil
	}

	end := filter.Offset + filter.Limit
	if end > total {
		end = total
	}

	return filtered[filter.Offset:end], total, nil
}

// FindLatestByClientID retrieves the most recent measurement for each type for a client.
func (r *InMemoryMeasurementRepository) FindLatestByClientID(_ context.Context, clientID string) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	latest := make(map[string]*entities.Measurement)
	for _, rec := range r.records {
		if rec.ClientID != clientID {
			continue
		}
		existing, ok := latest[rec.MeasurementType]
		if !ok || rec.MeasuredAt.After(existing.MeasuredAt) {
			latest[rec.MeasurementType] = rec
		}
	}

	result := make([]*entities.Measurement, 0, len(latest))
	for _, m := range latest {
		result = append(result, m)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].MeasurementType < result[j].MeasurementType
	})

	return result, nil
}
