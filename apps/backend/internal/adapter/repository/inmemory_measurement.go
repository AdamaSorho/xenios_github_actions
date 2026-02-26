package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryMeasurementRepository is an in-memory implementation of MeasurementRepository.
type InMemoryMeasurementRepository struct {
	mu           sync.RWMutex
	measurements []*entities.Measurement
	count        int
}

// NewInMemoryMeasurementRepository creates a new InMemoryMeasurementRepository.
func NewInMemoryMeasurementRepository() *InMemoryMeasurementRepository {
	return &InMemoryMeasurementRepository{}
}

// Add stores a measurement (for test setup, not part of the domain interface).
func (r *InMemoryMeasurementRepository) Add(m *entities.Measurement) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.count++
	stored := *m
	if stored.ID == "" {
		stored.ID = fmt.Sprintf("meas-%d", r.count)
	}
	if stored.CreatedAt.IsZero() {
		stored.CreatedAt = time.Now()
	}
	r.measurements = append(r.measurements, &stored)
}

// FindByClientID returns measurements for a given client with pagination.
func (r *InMemoryMeasurementRepository) FindByClientID(_ context.Context, clientID string, limit, offset int) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []*entities.Measurement
	for _, m := range r.measurements {
		if m.ClientID == clientID {
			matched = append(matched, m)
		}
	}

	return paginateMeasurements(matched, limit, offset), nil
}

// FindByClientIDAndType returns measurements for a client filtered by type and time window.
func (r *InMemoryMeasurementRepository) FindByClientIDAndType(_ context.Context, clientID string, measurementType string, since time.Time) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []*entities.Measurement
	for _, m := range r.measurements {
		if m.ClientID == clientID && m.MeasurementType == measurementType && !m.MeasuredAt.Before(since) {
			matched = append(matched, m)
		}
	}
	return matched, nil
}

// FindRecentByArtifactID returns measurements linked to a given artifact.
func (r *InMemoryMeasurementRepository) FindRecentByArtifactID(_ context.Context, artifactID string) ([]*entities.Measurement, error) {
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

func paginateMeasurements(items []*entities.Measurement, limit, offset int) []*entities.Measurement {
	if offset > len(items) {
		offset = len(items)
	}
	items = items[offset:]

	if limit <= 0 {
		limit = 50
	}
	if limit > len(items) {
		limit = len(items)
	}
	return items[:limit]
}
