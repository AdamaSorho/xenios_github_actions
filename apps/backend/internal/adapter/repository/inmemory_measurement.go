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

// Add stores a measurement (for testing/seeding).
func (r *InMemoryMeasurementRepository) Add(m *entities.Measurement) {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored := *m
	r.measurements = append(r.measurements, &stored)
}

// FindByClientID returns measurements for a client within a time range.
func (r *InMemoryMeasurementRepository) FindByClientID(_ context.Context, clientID string, from, to time.Time) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []*entities.Measurement
	for _, m := range r.measurements {
		if m.ClientID != clientID {
			continue
		}
		if !m.RecordedAt.Before(from) && !m.RecordedAt.After(to) {
			matched = append(matched, m)
		}
	}
	return matched, nil
}

// FindByArtifactID returns measurements extracted from a specific artifact.
func (r *InMemoryMeasurementRepository) FindByArtifactID(_ context.Context, artifactID string) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []*entities.Measurement
	for _, m := range r.measurements {
		if m.ArtifactID == artifactID {
			matched = append(matched, m)
		}
	}
	return matched, nil
}
