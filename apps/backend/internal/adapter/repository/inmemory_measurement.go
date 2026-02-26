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

// CreateBatch stores multiple measurements.
func (r *InMemoryMeasurementRepository) CreateBatch(_ context.Context, measurements []*entities.Measurement) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for _, m := range measurements {
		if m.ID == "" {
			b := make([]byte, 16)
			if _, err := rand.Read(b); err != nil {
				return fmt.Errorf("generate id: %w", err)
			}
			m.ID = hex.EncodeToString(b)
		}
		if m.CreatedAt.IsZero() {
			m.CreatedAt = now
		}
		stored := *m
		r.measurements = append(r.measurements, &stored)
	}
	return nil
}

// GetAll returns all stored measurements (for testing).
func (r *InMemoryMeasurementRepository) GetAll() []*entities.Measurement {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make([]*entities.Measurement, len(r.measurements))
	for i, m := range r.measurements {
		val := *m
		cp[i] = &val
	}
	return cp
}
