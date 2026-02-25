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
func (r *InMemoryMeasurementRepository) FindByClientID(_ context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*entities.Measurement
	for _, m := range r.records {
		if m.ClientID != filter.ClientID {
			continue
		}
		if filter.Type != "" && m.Type != filter.Type {
			continue
		}
		if filter.From != nil && m.MeasuredAt.Before(*filter.From) {
			continue
		}
		if filter.To != nil && m.MeasuredAt.After(*filter.To) {
			continue
		}
		filtered = append(filtered, m)
	}

	// Sort by measured_at descending
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].MeasuredAt.After(filtered[j].MeasuredAt)
	})

	total := len(filtered)

	// Apply pagination
	offset := filter.Offset
	if offset >= total {
		return []*entities.Measurement{}, total, nil
	}
	end := offset + filter.Limit
	if end > total {
		end = total
	}

	return filtered[offset:end], total, nil
}

// FindLatestByClientID returns the most recent measurement for each type.
func (r *InMemoryMeasurementRepository) FindLatestByClientID(_ context.Context, clientID string) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	latest := make(map[string]*entities.Measurement)
	for _, m := range r.records {
		if m.ClientID != clientID {
			continue
		}
		existing, exists := latest[m.Type]
		if !exists || m.MeasuredAt.After(existing.MeasuredAt) {
			latest[m.Type] = m
		}
	}

	result := make([]*entities.Measurement, 0, len(latest))
	for _, m := range latest {
		result = append(result, m)
	}
	return result, nil
}

// FindByType retrieves measurements for a client filtered by measurement type.
func (r *InMemoryMeasurementRepository) FindByType(_ context.Context, clientID, measurementType string) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entities.Measurement
	for _, m := range r.records {
		if m.ClientID == clientID && m.Type == measurementType {
			result = append(result, m)
		}
	}
	return result, nil
}
