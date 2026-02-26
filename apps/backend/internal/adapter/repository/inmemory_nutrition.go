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

// InMemoryNutritionRepository is an in-memory implementation of NutritionRepository.
type InMemoryNutritionRepository struct {
	mu           sync.RWMutex
	measurements []*entities.Measurement
	averages     []*entities.NutritionAverage
}

// NewInMemoryNutritionRepository creates a new InMemoryNutritionRepository.
func NewInMemoryNutritionRepository() *InMemoryNutritionRepository {
	return &InMemoryNutritionRepository{}
}

// BatchCreateMeasurements stores multiple nutrition measurements.
func (r *InMemoryNutritionRepository) BatchCreateMeasurements(_ context.Context, measurements []*entities.Measurement) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, m := range measurements {
		stored := *m
		if stored.ID == "" {
			stored.ID = generateID()
		}
		if stored.CreatedAt.IsZero() {
			stored.CreatedAt = time.Now()
		}
		r.measurements = append(r.measurements, &stored)
	}
	return nil
}

// StoreAverages stores computed nutrition averages.
func (r *InMemoryNutritionRepository) StoreAverages(_ context.Context, averages []*entities.NutritionAverage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, avg := range averages {
		stored := *avg
		if stored.ID == "" {
			stored.ID = generateID()
		}
		if stored.CreatedAt.IsZero() {
			stored.CreatedAt = time.Now()
		}
		r.averages = append(r.averages, &stored)
	}
	return nil
}

// FindMeasurementsByClientAndType retrieves measurements for a client filtered by type.
func (r *InMemoryNutritionRepository) FindMeasurementsByClientAndType(_ context.Context, clientID, measurementType string, limit int) ([]*entities.Measurement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entities.Measurement
	for _, m := range r.measurements {
		if m.ClientID == clientID && m.MeasurementType == measurementType {
			copy := *m
			result = append(result, &copy)
		}
	}

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

// FindAveragesByClient retrieves nutrition averages for a client.
func (r *InMemoryNutritionRepository) FindAveragesByClient(_ context.Context, clientID string) ([]*entities.NutritionAverage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entities.NutritionAverage
	for _, avg := range r.averages {
		if avg.ClientID == clientID {
			copy := *avg
			result = append(result, &copy)
		}
	}
	return result, nil
}

// GetMeasurements returns all stored measurements (for testing).
func (r *InMemoryNutritionRepository) GetMeasurements() []*entities.Measurement {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make([]*entities.Measurement, len(r.measurements))
	copy(cp, r.measurements)
	return cp
}

// GetAverages returns all stored averages (for testing).
func (r *InMemoryNutritionRepository) GetAverages() []*entities.NutritionAverage {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := make([]*entities.NutritionAverage, len(r.averages))
	copy(cp, r.averages)
	return cp
}

func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("id-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
