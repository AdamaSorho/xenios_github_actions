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
	return &InMemoryMeasurementRepository{
		measurements: make([]*entities.Measurement, 0),
	}
}

// BatchCreate stores multiple measurements.
func (r *InMemoryMeasurementRepository) BatchCreate(_ context.Context, measurements []*entities.Measurement) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for _, m := range measurements {
		stored := *m
		if stored.ID == "" {
			b := make([]byte, 16)
			if _, err := rand.Read(b); err != nil {
				return fmt.Errorf("generate id: %w", err)
			}
			stored.ID = hex.EncodeToString(b)
		}
		stored.CreatedAt = now
		r.measurements = append(r.measurements, &stored)
	}
	return nil
}

// FindByClientAndDateRange returns measurements for a client within a date range.
func (r *InMemoryMeasurementRepository) FindByClientAndDateRange(_ context.Context, clientID string, from, to time.Time) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entities.Measurement
	for _, m := range r.measurements {
		if m.ClientID == clientID &&
			!m.MeasuredAt.Before(from) &&
			!m.MeasuredAt.After(to) {
			copied := *m
			result = append(result, &copied)
		}
	}
	return result, nil
}

// GetAll returns all stored measurements (for testing).
func (r *InMemoryMeasurementRepository) GetAll() []*entities.Measurement {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*entities.Measurement, len(r.measurements))
	for i, m := range r.measurements {
		copied := *m
		result[i] = &copied
	}
	return result
}
