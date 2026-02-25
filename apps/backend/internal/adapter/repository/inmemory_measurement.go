package repository

import (
	"context"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryMeasurementRepository is an in-memory implementation of MeasurementRepository.
type InMemoryMeasurementRepository struct {
	mu           sync.RWMutex
	measurements []*entities.Measurement
}

// NewInMemoryMeasurementRepository creates a new InMemoryMeasurementRepository.
func NewInMemoryMeasurementRepository() *InMemoryMeasurementRepository {
	return &InMemoryMeasurementRepository{}
}

// Add inserts a measurement into the in-memory store (for testing).
func (r *InMemoryMeasurementRepository) Add(m *entities.Measurement) {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := *m
	r.measurements = append(r.measurements, &stored)
}

// FindRecentByClientID returns measurements for a client since a given time.
func (r *InMemoryMeasurementRepository) FindRecentByClientID(_ context.Context, clientID string, since time.Time) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entities.Measurement
	for _, m := range r.measurements {
		if m.ClientID == clientID && !m.MeasuredAt.Before(since) {
			cp := *m
			result = append(result, &cp)
		}
	}
	return result, nil
}

// FindByClientIDAndType returns measurements for a client filtered by type and time.
func (r *InMemoryMeasurementRepository) FindByClientIDAndType(_ context.Context, clientID string, measurementType string, since time.Time) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entities.Measurement
	for _, m := range r.measurements {
		if m.ClientID == clientID && m.MeasurementType == measurementType && !m.MeasuredAt.Before(since) {
			cp := *m
			result = append(result, &cp)
		}
	}
	return result, nil
}
