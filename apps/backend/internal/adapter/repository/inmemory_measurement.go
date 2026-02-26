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

// CreateBatch inserts multiple measurements in a single operation.
func (r *InMemoryMeasurementRepository) CreateBatch(_ context.Context, measurements []*entities.Measurement) ([]*entities.Measurement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var created []*entities.Measurement
	now := time.Now()

	for _, m := range measurements {
		stored := *m
		if stored.ID == "" {
			b := make([]byte, 16)
			if _, err := rand.Read(b); err != nil {
				return nil, fmt.Errorf("generate id: %w", err)
			}
			stored.ID = hex.EncodeToString(b)
		}
		if stored.CreatedAt.IsZero() {
			stored.CreatedAt = now
		}
		r.measurements = append(r.measurements, &stored)
		result := stored
		created = append(created, &result)
	}

	return created, nil
}

// FindByArtifactID retrieves all measurements linked to a specific artifact.
func (r *InMemoryMeasurementRepository) FindByArtifactID(_ context.Context, artifactID string) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entities.Measurement
	for _, m := range r.measurements {
		if m.ArtifactID == artifactID {
			cp := *m
			result = append(result, &cp)
		}
	}
	return result, nil
}

// GetAll returns all stored measurements (for testing).
func (r *InMemoryMeasurementRepository) GetAll() []*entities.Measurement {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make([]*entities.Measurement, len(r.measurements))
	for i, m := range r.measurements {
		v := *m
		cp[i] = &v
	}
	return cp
}
