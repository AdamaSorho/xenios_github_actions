package repository

import (
	"context"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryMeasurementRepository is a thread-safe in-memory implementation
// of MeasurementRepository for testing.
type InMemoryMeasurementRepository struct {
	mu           sync.RWMutex
	measurements []entities.Measurement
}

// NewInMemoryMeasurementRepository creates a new in-memory measurement repository.
func NewInMemoryMeasurementRepository() *InMemoryMeasurementRepository {
	return &InMemoryMeasurementRepository{}
}

// UpsertBatch inserts measurements, skipping duplicates (same client_id, measurement_type, measured_at, source).
func (r *InMemoryMeasurementRepository) UpsertBatch(_ context.Context, measurements []entities.Measurement) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	inserted := 0
	for _, m := range measurements {
		if !r.isDuplicate(m) {
			r.measurements = append(r.measurements, m)
			inserted++
		}
	}
	return inserted, nil
}

// Average computes the average value of a measurement type for a client within a time window.
func (r *InMemoryMeasurementRepository) Average(_ context.Context, clientID string, mt entities.MeasurementType, source entities.WearableSource, since time.Time) (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sum float64
	var count int
	for _, m := range r.measurements {
		if m.ClientID == clientID && m.MeasurementType == mt && m.Source == source && !m.MeasuredAt.Before(since) {
			sum += m.Value
			count++
		}
	}
	if count == 0 {
		return 0, nil
	}
	return sum / float64(count), nil
}

// GetMeasurements returns all stored measurements (for test assertions).
func (r *InMemoryMeasurementRepository) GetMeasurements() []entities.Measurement {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]entities.Measurement, len(r.measurements))
	copy(out, r.measurements)
	return out
}

func (r *InMemoryMeasurementRepository) isDuplicate(m entities.Measurement) bool {
	for _, existing := range r.measurements {
		if existing.ClientID == m.ClientID &&
			existing.MeasurementType == m.MeasurementType &&
			existing.MeasuredAt.Equal(m.MeasuredAt) &&
			existing.Source == m.Source {
			return true
		}
	}
	return false
}
