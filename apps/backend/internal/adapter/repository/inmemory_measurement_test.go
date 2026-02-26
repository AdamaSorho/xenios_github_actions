package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryMeasurementRepository_CreateBatch_StoresAll(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()

	measurements := []*entities.Measurement{
		{ClientID: "c1", RecordedBy: "r1", MeasurementType: entities.MeasurementTypeWeight, Value: 85.5, Unit: "kg", MeasuredAt: now},
		{ClientID: "c1", RecordedBy: "r1", MeasurementType: entities.MeasurementTypeBodyFatPct, Value: 22.3, Unit: "%", MeasuredAt: now},
	}

	created, err := repo.CreateBatch(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(created) != 2 {
		t.Fatalf("expected 2 created, got %d", len(created))
	}
	for _, m := range created {
		if m.ID == "" {
			t.Error("expected non-empty ID")
		}
	}
}

func TestInMemoryMeasurementRepository_CreateBatch_GeneratesIDs(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	measurements := []*entities.Measurement{
		{ClientID: "c1", RecordedBy: "r1", MeasurementType: entities.MeasurementTypeWeight, Value: 85.5, Unit: "kg", MeasuredAt: time.Now()},
	}

	created, err := repo.CreateBatch(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created[0].ID == "" {
		t.Error("expected ID to be generated")
	}
	if created[0].CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestInMemoryMeasurementRepository_CreateBatch_PreservesExistingID(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	measurements := []*entities.Measurement{
		{ID: "custom-id", ClientID: "c1", RecordedBy: "r1", MeasurementType: entities.MeasurementTypeWeight, Value: 85.5, Unit: "kg", MeasuredAt: time.Now()},
	}

	created, err := repo.CreateBatch(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created[0].ID != "custom-id" {
		t.Errorf("expected ID 'custom-id', got %q", created[0].ID)
	}
}

func TestInMemoryMeasurementRepository_FindByArtifactID_ReturnsMatching(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()

	measurements := []*entities.Measurement{
		{ClientID: "c1", RecordedBy: "r1", MeasurementType: entities.MeasurementTypeWeight, Value: 85.5, Unit: "kg", MeasuredAt: now, ArtifactID: "art-1"},
		{ClientID: "c1", RecordedBy: "r1", MeasurementType: entities.MeasurementTypeBMR, Value: 1847, Unit: "kcal", MeasuredAt: now, ArtifactID: "art-1"},
		{ClientID: "c1", RecordedBy: "r1", MeasurementType: entities.MeasurementTypeWeight, Value: 90.0, Unit: "kg", MeasuredAt: now, ArtifactID: "art-2"},
	}

	_, err := repo.CreateBatch(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, err := repo.FindByArtifactID(context.Background(), "art-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("expected 2 measurements for art-1, got %d", len(found))
	}
}

func TestInMemoryMeasurementRepository_FindByArtifactID_NoMatch_ReturnsEmpty(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	found, err := repo.FindByArtifactID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(found) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(found))
	}
}

func TestInMemoryMeasurementRepository_GetAll_ReturnsAll(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()

	measurements := []*entities.Measurement{
		{ClientID: "c1", RecordedBy: "r1", MeasurementType: entities.MeasurementTypeWeight, Value: 85.5, Unit: "kg", MeasuredAt: now},
		{ClientID: "c2", RecordedBy: "r2", MeasurementType: entities.MeasurementTypeBMR, Value: 1847, Unit: "kcal", MeasuredAt: now},
	}

	_, err := repo.CreateBatch(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all := repo.GetAll()
	if len(all) != 2 {
		t.Fatalf("expected 2 measurements, got %d", len(all))
	}
}

func TestInMemoryMeasurementRepository_CreateBatch_Empty_ReturnsNil(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	created, err := repo.CreateBatch(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(created) != 0 {
		t.Errorf("expected 0 created, got %d", len(created))
	}
}
