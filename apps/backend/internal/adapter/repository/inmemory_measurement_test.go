package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryMeasurementRepository_FindByClientID_FiltersCorrectly(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()

	repo.Add(&entities.Measurement{
		ID: "m-1", ClientID: "client-1", MarkerName: "LDL",
		RecordedAt: now.AddDate(0, 0, -5), CreatedAt: now,
	})
	repo.Add(&entities.Measurement{
		ID: "m-2", ClientID: "client-2", MarkerName: "HDL",
		RecordedAt: now.AddDate(0, 0, -3), CreatedAt: now,
	})
	repo.Add(&entities.Measurement{
		ID: "m-3", ClientID: "client-1", MarkerName: "Glucose",
		RecordedAt: now.AddDate(0, 0, -1), CreatedAt: now,
	})

	results, err := repo.FindByClientID(context.Background(), "client-1", now.AddDate(0, 0, -7), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 measurements, got %d", len(results))
	}
}

func TestInMemoryMeasurementRepository_FindByClientID_TimeRange(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()

	repo.Add(&entities.Measurement{
		ID: "m-1", ClientID: "client-1", MarkerName: "LDL",
		RecordedAt: now.AddDate(0, 0, -30), CreatedAt: now,
	})
	repo.Add(&entities.Measurement{
		ID: "m-2", ClientID: "client-1", MarkerName: "HDL",
		RecordedAt: now.AddDate(0, 0, -3), CreatedAt: now,
	})

	// Only look at last 7 days
	results, err := repo.FindByClientID(context.Background(), "client-1", now.AddDate(0, 0, -7), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 measurement within 7-day range, got %d", len(results))
	}
}

func TestInMemoryMeasurementRepository_FindByArtifactID_FiltersCorrectly(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()

	repo.Add(&entities.Measurement{
		ID: "m-1", ArtifactID: "artifact-1", MarkerName: "LDL",
		RecordedAt: now, CreatedAt: now,
	})
	repo.Add(&entities.Measurement{
		ID: "m-2", ArtifactID: "artifact-2", MarkerName: "HDL",
		RecordedAt: now, CreatedAt: now,
	})
	repo.Add(&entities.Measurement{
		ID: "m-3", ArtifactID: "artifact-1", MarkerName: "Glucose",
		RecordedAt: now, CreatedAt: now,
	})

	results, err := repo.FindByArtifactID(context.Background(), "artifact-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 measurements for artifact-1, got %d", len(results))
	}
}

func TestInMemoryMeasurementRepository_FindByArtifactID_NotFound(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	results, err := repo.FindByArtifactID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 measurements, got %d", len(results))
	}
}
