package repository

import (
	"context"
	"sync"

	domainrepo "github.com/xenios/backend/internal/domain/repository"
)

// InMemoryMeasurementRepository is an in-memory implementation of MeasurementRepository.
type InMemoryMeasurementRepository struct {
	mu           sync.RWMutex
	measurements []domainrepo.MeasurementInput
}

// NewInMemoryMeasurementRepository creates a new InMemoryMeasurementRepository.
func NewInMemoryMeasurementRepository() *InMemoryMeasurementRepository {
	return &InMemoryMeasurementRepository{
		measurements: make([]domainrepo.MeasurementInput, 0),
	}
}

// CreateBatch stores multiple measurements atomically.
func (r *InMemoryMeasurementRepository) CreateBatch(_ context.Context, inputs []domainrepo.MeasurementInput) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.measurements = append(r.measurements, inputs...)
	return len(inputs), nil
}

// All returns all stored measurements (test helper).
func (r *InMemoryMeasurementRepository) All() []domainrepo.MeasurementInput {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domainrepo.MeasurementInput, len(r.measurements))
	copy(result, r.measurements)
	return result
}
