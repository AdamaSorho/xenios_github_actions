package repository

import (
	"context"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryMeasurementRepository is a test-friendly in-memory implementation of MeasurementRepository.
type InMemoryMeasurementRepository struct {
	mu           sync.RWMutex
	measurements []entities.Measurement
}

// NewInMemoryMeasurementRepository creates a new in-memory measurement store.
func NewInMemoryMeasurementRepository() *InMemoryMeasurementRepository {
	return &InMemoryMeasurementRepository{}
}

func (r *InMemoryMeasurementRepository) UpsertBatch(_ context.Context, measurements []entities.Measurement) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	inserted := 0
	for _, m := range measurements {
		if r.isDuplicate(m) {
			continue
		}
		r.measurements = append(r.measurements, m)
		inserted++
	}
	return inserted, nil
}

func (r *InMemoryMeasurementRepository) isDuplicate(m entities.Measurement) bool {
	for _, existing := range r.measurements {
		if existing.ClientID == m.ClientID &&
			existing.Source == m.Source &&
			existing.MeasurementType == m.MeasurementType &&
			existing.MeasuredAt.Equal(m.MeasuredAt) {
			return true
		}
	}
	return false
}

func (r *InMemoryMeasurementRepository) Average(_ context.Context, clientID string, source entities.WearableSource, mt entities.MeasurementType, since time.Time) (*float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sum float64
	var count int
	for _, m := range r.measurements {
		if m.ClientID == clientID && m.Source == source && m.MeasurementType == mt && !m.MeasuredAt.Before(since) {
			sum += m.Value
			count++
		}
	}
	if count == 0 {
		return nil, nil
	}
	avg := sum / float64(count)
	return &avg, nil
}

// All returns all stored measurements (for testing).
func (r *InMemoryMeasurementRepository) All() []entities.Measurement {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]entities.Measurement, len(r.measurements))
	copy(out, r.measurements)
	return out
}
