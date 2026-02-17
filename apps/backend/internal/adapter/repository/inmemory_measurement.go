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

// CreateBatch stores multiple measurements.
func (r *InMemoryMeasurementRepository) CreateBatch(_ context.Context, measurements []*entities.Measurement) ([]*entities.Measurement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	results := make([]*entities.Measurement, len(measurements))
	for i, m := range measurements {
		stored := *m
		if stored.ID == "" {
			b := make([]byte, 16)
			if _, err := rand.Read(b); err != nil {
				return nil, fmt.Errorf("generate id: %w", err)
			}
			stored.ID = hex.EncodeToString(b)
		}
		if stored.CreatedAt.IsZero() {
			stored.CreatedAt = time.Now()
		}
		r.measurements = append(r.measurements, &stored)
		result := stored
		results[i] = &result
	}

	return results, nil
}

// FindByArtifactID retrieves all measurements linked to a specific artifact.
func (r *InMemoryMeasurementRepository) FindByArtifactID(_ context.Context, artifactID string) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*entities.Measurement
	for _, m := range r.measurements {
		if m.ArtifactID == artifactID {
			copy := *m
			results = append(results, &copy)
		}
	}
	return results, nil
}

// FindByClientID retrieves measurements for a client.
func (r *InMemoryMeasurementRepository) FindByClientID(_ context.Context, clientID string) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*entities.Measurement
	for _, m := range r.measurements {
		if m.ClientID == clientID {
			copy := *m
			results = append(results, &copy)
		}
	}
	return results, nil
}

// GetAll returns all stored measurements (for testing).
func (r *InMemoryMeasurementRepository) GetAll() []*entities.Measurement {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make([]*entities.Measurement, len(r.measurements))
	for i, m := range r.measurements {
		copy := *m
		results[i] = &copy
	}
	return results
}
