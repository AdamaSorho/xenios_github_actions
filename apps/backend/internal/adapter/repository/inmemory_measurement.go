package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryMeasurementRepository is an in-memory implementation of MeasurementRepository.
type InMemoryMeasurementRepository struct {
	mu           sync.RWMutex
	measurements []entities.WearableMeasurement
}

// NewInMemoryMeasurementRepository creates a new InMemoryMeasurementRepository.
func NewInMemoryMeasurementRepository() *InMemoryMeasurementRepository {
	return &InMemoryMeasurementRepository{}
}

// BulkUpsert inserts measurements, skipping duplicates for the same client+source+type+date.
func (r *InMemoryMeasurementRepository) BulkUpsert(_ context.Context, measurements []entities.WearableMeasurement) (int, error) {
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

// GetAverages returns the average value for a measurement type within a time window.
func (r *InMemoryMeasurementRepository) GetAverages(_ context.Context, clientID string, source entities.WearableSource, measurementType entities.MeasurementType, from time.Time, to time.Time) (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sum float64
	var count int
	for _, m := range r.measurements {
		if m.ClientID != clientID {
			continue
		}
		if m.Source != source {
			continue
		}
		if m.MeasurementType != measurementType {
			continue
		}
		if m.MeasuredAt.Before(from) || m.MeasuredAt.After(to) {
			continue
		}
		sum += m.Value
		count++
	}

	if count == 0 {
		return 0, fmt.Errorf("no measurements found")
	}
	return sum / float64(count), nil
}

// GetAll returns all stored measurements (for testing).
func (r *InMemoryMeasurementRepository) GetAll() []entities.WearableMeasurement {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make([]entities.WearableMeasurement, len(r.measurements))
	copy(cp, r.measurements)
	return cp
}

// Count returns the number of stored measurements (for testing).
func (r *InMemoryMeasurementRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.measurements)
}

func (r *InMemoryMeasurementRepository) isDuplicate(m entities.WearableMeasurement) bool {
	dateKey := m.MeasuredAt.Format("2006-01-02")
	for _, existing := range r.measurements {
		if existing.ClientID == m.ClientID &&
			existing.Source == m.Source &&
			existing.MeasurementType == m.MeasurementType &&
			existing.MeasuredAt.Format("2006-01-02") == dateKey {
			return true
		}
	}
	return false
}
