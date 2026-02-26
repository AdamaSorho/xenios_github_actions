package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

// mockMeasurementRepo verifies the interface can be implemented.
type mockMeasurementRepo struct{}

func (m *mockMeasurementRepo) CreateBatch(_ context.Context, measurements []*entities.Measurement) ([]*entities.Measurement, error) {
	return measurements, nil
}

func (m *mockMeasurementRepo) FindByArtifactID(_ context.Context, _ string) ([]*entities.Measurement, error) {
	return nil, nil
}

// Compile-time interface check.
var _ MeasurementRepository = &mockMeasurementRepo{}

func TestMeasurementRepository_InterfaceCompiles(t *testing.T) {
	var repo MeasurementRepository = &mockMeasurementRepo{}
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}
