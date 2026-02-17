package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryMeasurementRepository_Create_GeneratesID(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	m := &entities.Measurement{
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "fasting_glucose",
		Value:           98,
		Unit:            "mg/dL",
	}

	created, err := repo.Create(context.Background(), m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created.ID == "" {
		t.Error("expected generated ID")
	}
}

func TestInMemoryMeasurementRepository_Create_SetsTimestamps(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	m := &entities.Measurement{
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "fasting_glucose",
		Value:           98,
		Unit:            "mg/dL",
	}

	created, err := repo.Create(context.Background(), m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if created.MeasuredAt.IsZero() {
		t.Error("expected MeasuredAt to be set")
	}
}

func TestInMemoryMeasurementRepository_BatchCreate_CreatesAll(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	measurements := []*entities.Measurement{
		{ClientID: "c1", RecordedBy: "r1", MeasurementType: "fasting_glucose", Value: 98, Unit: "mg/dL"},
		{ClientID: "c1", RecordedBy: "r1", MeasurementType: "ldl_cholesterol", Value: 142, Unit: "mg/dL"},
		{ClientID: "c1", RecordedBy: "r1", MeasurementType: "hdl_cholesterol", Value: 55, Unit: "mg/dL"},
	}

	created, err := repo.BatchCreate(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(created) != 3 {
		t.Fatalf("expected 3 created measurements, got %d", len(created))
	}

	for _, c := range created {
		if c.ID == "" {
			t.Error("expected all measurements to have generated IDs")
		}
	}
}

func TestInMemoryMeasurementRepository_FindByArtifactID_ReturnsMatching(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	artifactID := "artifact-123"

	m1 := &entities.Measurement{
		ClientID: "c1", RecordedBy: "r1", MeasurementType: "fasting_glucose",
		Value: 98, Unit: "mg/dL", ArtifactID: &artifactID,
	}
	m2 := &entities.Measurement{
		ClientID: "c1", RecordedBy: "r1", MeasurementType: "ldl_cholesterol",
		Value: 142, Unit: "mg/dL", ArtifactID: &artifactID,
	}
	otherArtifact := "artifact-999"
	m3 := &entities.Measurement{
		ClientID: "c1", RecordedBy: "r1", MeasurementType: "hdl_cholesterol",
		Value: 55, Unit: "mg/dL", ArtifactID: &otherArtifact,
	}

	repo.Create(context.Background(), m1)
	repo.Create(context.Background(), m2)
	repo.Create(context.Background(), m3)

	results, err := repo.FindByArtifactID(context.Background(), artifactID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results for artifact-123, got %d", len(results))
	}
}

func TestInMemoryMeasurementRepository_FindByArtifactID_NoResults_ReturnsEmpty(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	results, err := repo.FindByArtifactID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
