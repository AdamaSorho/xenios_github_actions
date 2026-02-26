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
	offset := (filter.Page - 1) * filter.Limit
	if offset >= total {
		return []*entities.Measurement{}, total, nil
	}

	end := offset + filter.Limit
	if end > total {
		end = total
	}

	return filtered[offset:end], total, nil
}

// FindLatestByClientID returns the most recent measurement for each type for a client.
func (r *InMemoryMeasurementRepository) FindLatestByClientID(_ context.Context, clientID string) ([]*entities.LatestMeasurement, error) {
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

	result := make([]*entities.LatestMeasurement, 0, len(latest))
	for _, m := range latest {
		result = append(result, &entities.LatestMeasurement{
			MeasurementType: m.MeasurementType,
			Value:           m.Value,
			Unit:            m.Unit,
			MeasuredAt:      m.MeasuredAt,
			Flag:            m.Flag,
			ReferenceLow:    m.ReferenceLow,
			ReferenceHigh:   m.ReferenceHigh,
		})
	}

	return result, nil
}

// FindByType retrieves measurements for a client filtered by measurement type.
func (r *InMemoryMeasurementRepository) FindByType(_ context.Context, clientID, measurementType string) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entities.Measurement
	for _, rec := range r.records {
		if rec.ClientID == clientID && rec.MeasurementType == measurementType {
			result = append(result, rec)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].MeasuredAt.After(result[j].MeasuredAt)
	})

	return result, nil
}

// Seed adds a measurement directly for testing purposes.
func (r *InMemoryMeasurementRepository) Seed(m *entities.Measurement) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, m)
}
