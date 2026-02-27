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
	measurements map[string]*entities.Measurement
}

// NewInMemoryMeasurementRepository creates a new InMemoryMeasurementRepository.
func NewInMemoryMeasurementRepository() *InMemoryMeasurementRepository {
	return &InMemoryMeasurementRepository{
		measurements: make(map[string]*entities.Measurement),
	}
}

// CreateBatch stores multiple measurements atomically.
func (r *InMemoryMeasurementRepository) CreateBatch(_ context.Context, measurements []*entities.Measurement) ([]*entities.Measurement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]*entities.Measurement, 0, len(measurements))
	for _, m := range measurements {
		stored := *m
		if stored.ID == "" {
			b := make([]byte, 16)
			if _, err := rand.Read(b); err != nil {
				return nil, fmt.Errorf("generate id: %w", err)
			}
			stored.ID = hex.EncodeToString(b)
		}
		stored.CreatedAt = time.Now()
		r.measurements[stored.ID] = &stored
		copy := stored
		result = append(result, &copy)
	}
	return result, nil
}

// FindByArtifactID returns all measurements linked to a given artifact.
func (r *InMemoryMeasurementRepository) FindByArtifactID(_ context.Context, artifactID string) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entities.Measurement
	for _, m := range r.measurements {
		if m.ArtifactID != nil && *m.ArtifactID == artifactID {
			copy := *m
			result = append(result, &copy)
		}
	}
	return result, nil
}

// GetAll returns all stored measurements (test helper).
func (r *InMemoryMeasurementRepository) GetAll() []*entities.Measurement {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*entities.Measurement, 0, len(r.measurements))
	for _, m := range r.measurements {
		copy := *m
		result = append(result, &copy)
	}
	return result
}
