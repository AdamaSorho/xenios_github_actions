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

// Create stores a new measurement.
func (r *InMemoryMeasurementRepository) Create(_ context.Context, m *entities.Measurement) (*entities.Measurement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if m.ID == "" {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			return nil, fmt.Errorf("generate id: %w", err)
		}
		m.ID = hex.EncodeToString(b)
	}

	now := time.Now()
	m.CreatedAt = now
	if m.MeasuredAt.IsZero() {
		m.MeasuredAt = now
	}

	stored := *m
	r.measurements[stored.ID] = &stored
	result := stored
	return &result, nil
}

// BatchCreate stores multiple measurements.
func (r *InMemoryMeasurementRepository) BatchCreate(ctx context.Context, measurements []*entities.Measurement) ([]*entities.Measurement, error) {
	var created []*entities.Measurement
	for _, m := range measurements {
		result, err := r.Create(ctx, m)
		if err != nil {
			return nil, err
		}
		created = append(created, result)
	}
	return created, nil
}

// FindByArtifactID returns all measurements linked to a given artifact.
func (r *InMemoryMeasurementRepository) FindByArtifactID(_ context.Context, artifactID string) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*entities.Measurement
	for _, m := range r.measurements {
		if m.ArtifactID != nil && *m.ArtifactID == artifactID {
			copy := *m
			results = append(results, &copy)
		}
	}
	return results, nil
}
