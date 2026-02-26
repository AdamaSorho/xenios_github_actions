package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
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

// CreateBatch stores a batch of measurements.
func (r *InMemoryMeasurementRepository) CreateBatch(_ context.Context, measurements []*entities.Measurement) ([]*entities.Measurement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]*entities.Measurement, len(measurements))
	now := time.Now()

	for i, m := range measurements {
		stored := *m
		if stored.ID == "" {
			b := make([]byte, 16)
			if _, err := rand.Read(b); err != nil {
				return nil, fmt.Errorf("generate id: %w", err)
			}
			stored.ID = hex.EncodeToString(b)
		}
		stored.CreatedAt = now
		r.measurements = append(r.measurements, &stored)
		cp := stored
		result[i] = &cp
	}

	return result, nil
}

// GetAll returns all stored measurements (for testing).
func (r *InMemoryMeasurementRepository) GetAll() []*entities.Measurement {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make([]*entities.Measurement, len(r.measurements))
	for i, m := range r.measurements {
		stored := *m
		cp[i] = &stored
	}
	return cp
}
