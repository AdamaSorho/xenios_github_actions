package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/repository"
)

// InMemoryMeasurementRepository is an in-memory implementation of MeasurementRepository for testing.
type InMemoryMeasurementRepository struct {
	mu           sync.RWMutex
	measurements []repository.Measurement
	count        int
}

// NewInMemoryMeasurementRepository creates a new InMemoryMeasurementRepository.
func NewInMemoryMeasurementRepository() *InMemoryMeasurementRepository {
	return &InMemoryMeasurementRepository{}
}

// BatchCreate stores multiple measurements.
func (r *InMemoryMeasurementRepository) BatchCreate(_ context.Context, measurements []repository.Measurement) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for _, m := range measurements {
		r.count++
		stored := m
		if stored.ID == "" {
			stored.ID = fmt.Sprintf("meas-%d", r.count)
		}
		if stored.CreatedAt.IsZero() {
			stored.CreatedAt = now
		}
		r.measurements = append(r.measurements, stored)
	}
	return nil
}

// GetAll returns all stored measurements (for testing).
func (r *InMemoryMeasurementRepository) GetAll() []repository.Measurement {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make([]repository.Measurement, len(r.measurements))
	copy(cp, r.measurements)
	return cp
}
