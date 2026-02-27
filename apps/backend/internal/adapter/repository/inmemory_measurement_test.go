package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryMeasurementRepository_CreateBatch_StoresMeasurements(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	artifactID := "artifact-1"

	measurements := []*entities.Measurement{
		{
			ClientID:        "client-1",
			RecordedBy:      "coach-1",
			MeasurementType: entities.MeasurementTypeWeight,
			Value:           85.0,
			Unit:            "kg",
			MeasuredAt:      time.Now(),
			ArtifactID:      &artifactID,
		},
		{
			ClientID:        "client-1",
			RecordedBy:      "coach-1",
			MeasurementType: entities.MeasurementTypeBodyFatPct,
			Value:           20.5,
			Unit:            "%",
			MeasuredAt:      time.Now(),
			ArtifactID:      &artifactID,
		},
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

func TestInMemoryMeasurementRepository_FindByArtifactID_ReturnsMatching(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	artifactID := "artifact-1"
	otherArtifactID := "artifact-2"

	_, _ = repo.CreateBatch(context.Background(), []*entities.Measurement{
		{ClientID: "c1", RecordedBy: "r1", MeasurementType: entities.MeasurementTypeWeight, Value: 85, Unit: "kg", MeasuredAt: time.Now(), ArtifactID: &artifactID},
		{ClientID: "c1", RecordedBy: "r1", MeasurementType: entities.MeasurementTypeBMR, Value: 1800, Unit: "kcal", MeasuredAt: time.Now(), ArtifactID: &otherArtifactID},
	})

	found, err := repo.FindByArtifactID(context.Background(), artifactID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(found) != 1 {
		t.Fatalf("expected 1 measurement, got %d", len(found))
	}
	if found[0].MeasurementType != entities.MeasurementTypeWeight {
		t.Errorf("expected weight, got %s", found[0].MeasurementType)
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

func TestInMemoryMeasurementRepository_CreateBatch_Empty_ReturnsEmpty(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	created, err := repo.CreateBatch(context.Background(), []*entities.Measurement{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(created) != 0 {
		t.Errorf("expected 0 created, got %d", len(created))
	}
}
