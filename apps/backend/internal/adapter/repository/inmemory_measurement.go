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

	m.ID = generateRepoID()
	m.CreatedAt = time.Now()
	r.records = append(r.records, m)
	return m, nil
}

// FindByClientID retrieves measurements for a client with filtering and pagination.
func (r *InMemoryMeasurementRepository) FindByClientID(_ context.Context, filter entities.MeasurementFilter) (*entities.MeasurementPage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*entities.Measurement
	for _, rec := range r.records {
		if rec.ClientID != filter.ClientID {
			continue
		}
		if filter.Type != "" && rec.Type != filter.Type {
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
	offset := (filter.Page - 1) * filter.Limit
	if offset >= total {
		return &entities.MeasurementPage{
			Measurements: []*entities.Measurement{},
			Page:         filter.Page,
			Limit:        filter.Limit,
			Total:        total,
		}, nil
	}

	end := offset + filter.Limit
	if end > total {
		end = total
	}

	return &entities.MeasurementPage{
		Measurements: filtered[offset:end],
		Page:         filter.Page,
		Limit:        filter.Limit,
		Total:        total,
	}, nil
}

// FindLatestByClientID returns the most recent measurement for each type.
func (r *InMemoryMeasurementRepository) FindLatestByClientID(_ context.Context, clientID string) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	latest := make(map[string]*entities.Measurement)
	for _, rec := range r.records {
		if rec.ClientID != clientID {
			continue
		}
		existing, ok := latest[rec.Type]
		if !ok || rec.MeasuredAt.After(existing.MeasuredAt) {
			latest[rec.Type] = rec
		}
	}

	result := make([]*entities.Measurement, 0, len(latest))
	for _, m := range latest {
		result = append(result, m)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Type < result[j].Type
	})

	return result, nil
}

// FindByType retrieves measurements of a specific type for a client.
func (r *InMemoryMeasurementRepository) FindByType(_ context.Context, clientID, measurementType string) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entities.Measurement
	for _, rec := range r.records {
		if rec.ClientID == clientID && rec.Type == measurementType {
			result = append(result, rec)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].MeasuredAt.After(result[j].MeasuredAt)
	})

	return result, nil
}
