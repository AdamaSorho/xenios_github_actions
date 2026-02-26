package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	domainrepo "github.com/xenios/backend/internal/domain/repository"
)

// Compile-time interface check.
var _ domainrepo.MeasurementRepository = &InMemoryMeasurementRepository{}

func TestInMemoryMeasurementRepository_CreateBatch_StoresMeasurements(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()
	measurements := []*entities.Measurement{
		{
			ClientID:         "client-1",
			RecordedBy:       "coach-1",
			MeasurementType:  entities.MeasurementTypeWeight,
			Value:            85.4,
			Unit:             "kg",
			MeasuredAt:       now,
			ArtifactID:       "artifact-1",
			ExtractionStatus: entities.ExtractionStatusComplete,
		},
		{
			ClientID:         "client-1",
			RecordedBy:       "coach-1",
			MeasurementType:  entities.MeasurementTypeBodyFatPct,
			Value:            22.3,
			Unit:             "%",
			MeasuredAt:       now,
			ArtifactID:       "artifact-1",
			ExtractionStatus: entities.ExtractionStatusComplete,
		},
	}

	result, err := repo.CreateBatch(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 measurements, got %d", len(result))
	}
	for _, m := range result {
		if m.ID == "" {
			t.Error("expected measurement to have an ID")
		}
	}
}

func TestInMemoryMeasurementRepository_CreateBatch_Empty_ReturnsNil(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	result, err := repo.CreateBatch(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestInMemoryMeasurementRepository_FindByArtifactID_Found(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()
	measurements := []*entities.Measurement{
		{
			ClientID:        "client-1",
			RecordedBy:      "coach-1",
			MeasurementType: entities.MeasurementTypeWeight,
			Value:           85.4,
			Unit:            "kg",
			MeasuredAt:      now,
			ArtifactID:      "artifact-1",
		},
		{
			ClientID:        "client-1",
			RecordedBy:      "coach-1",
			MeasurementType: entities.MeasurementTypeBMR,
			Value:           1847,
			Unit:            "kcal",
			MeasuredAt:      now,
			ArtifactID:      "artifact-2",
		},
	}
	_, err := repo.CreateBatch(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, err := repo.FindByArtifactID(context.Background(), "artifact-1")
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

func TestInMemoryMeasurementRepository_FindByArtifactID_NotFound(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	found, err := repo.FindByArtifactID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(found) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(found))
	}
}

func TestInMemoryMeasurementRepository_All_ReturnsAllMeasurements(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()
	_, _ = repo.CreateBatch(context.Background(), []*entities.Measurement{
		{
			ClientID:        "client-1",
			RecordedBy:      "coach-1",
			MeasurementType: entities.MeasurementTypeWeight,
			Value:           85.4,
			Unit:            "kg",
			MeasuredAt:      now,
			ArtifactID:      "artifact-1",
		},
	})
	all := repo.All()
	if len(all) != 1 {
		t.Errorf("expected 1 measurement, got %d", len(all))
	}
}
