package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryMeasurementRepository_CreateBatch_StoresMeasurements(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	ctx := context.Background()

	measurements := []*entities.Measurement{
		{
			ClientID:        "client-1",
			RecordedBy:      "coach-1",
			MeasurementType: entities.MeasurementTypeWeight,
			Value:           85.5,
			Unit:            "kg",
			MeasuredAt:      time.Now(),
			ArtifactID:      "art-1",
		},
		{
			ClientID:        "client-1",
			RecordedBy:      "coach-1",
			MeasurementType: entities.MeasurementTypeBodyFatPct,
			Value:           22.3,
			Unit:            "%",
			MeasuredAt:      time.Now(),
			ArtifactID:      "art-1",
		},
	}

	results, err := repo.CreateBatch(ctx, measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.ID == "" {
			t.Error("expected generated ID")
		}
	}
}

func TestInMemoryMeasurementRepository_FindByArtifactID_ReturnsMatching(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	ctx := context.Background()

	measurements := []*entities.Measurement{
		{
			ClientID:        "client-1",
			RecordedBy:      "coach-1",
			MeasurementType: entities.MeasurementTypeWeight,
			Value:           85.5,
			Unit:            "kg",
			MeasuredAt:      time.Now(),
			ArtifactID:      "art-1",
		},
		{
			ClientID:        "client-1",
			RecordedBy:      "coach-1",
			MeasurementType: entities.MeasurementTypeBMR,
			Value:           1800,
			Unit:            "kcal",
			MeasuredAt:      time.Now(),
			ArtifactID:      "art-2",
		},
	}

	_, err := repo.CreateBatch(ctx, measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results, err := repo.FindByArtifactID(ctx, "art-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ArtifactID != "art-1" {
		t.Errorf("expected artifact_id 'art-1', got %q", results[0].ArtifactID)
	}
}

func TestInMemoryMeasurementRepository_FindByClientID_ReturnsMatching(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	ctx := context.Background()

	measurements := []*entities.Measurement{
		{
			ClientID:        "client-1",
			RecordedBy:      "coach-1",
			MeasurementType: entities.MeasurementTypeWeight,
			Value:           85.5,
			Unit:            "kg",
			MeasuredAt:      time.Now(),
			ArtifactID:      "art-1",
		},
		{
			ClientID:        "client-2",
			RecordedBy:      "coach-1",
			MeasurementType: entities.MeasurementTypeWeight,
			Value:           70.0,
			Unit:            "kg",
			MeasuredAt:      time.Now(),
			ArtifactID:      "art-2",
		},
	}

	_, err := repo.CreateBatch(ctx, measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results, err := repo.FindByClientID(ctx, "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ClientID != "client-1" {
		t.Errorf("expected client_id 'client-1', got %q", results[0].ClientID)
	}
}

func TestInMemoryMeasurementRepository_FindByArtifactID_NoResults_ReturnsEmpty(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	ctx := context.Background()

	results, err := repo.FindByArtifactID(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil && len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}
