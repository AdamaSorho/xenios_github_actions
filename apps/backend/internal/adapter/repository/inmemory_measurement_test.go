package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	domainrepo "github.com/xenios/backend/internal/domain/repository"
)

// Compile-time interface compliance check.
var _ domainrepo.MeasurementRepository = &InMemoryMeasurementRepository{}

func TestInMemoryMeasurementRepository_Add_StoresMeasurement(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	repo.Add(&entities.Measurement{
		ClientID:        "client-1",
		MeasurementType: "ldl",
		Value:           190,
		Unit:            "mg/dL",
	})

	results, err := repo.FindByClientID(context.Background(), "client-1", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 measurement, got %d", len(results))
	}
	if results[0].Value != 190 {
		t.Errorf("expected value 190, got %f", results[0].Value)
	}
}

func TestInMemoryMeasurementRepository_Add_AssignsIDAndTimestamp(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	repo.Add(&entities.Measurement{
		ClientID:        "client-1",
		MeasurementType: "ldl",
		Value:           190,
		Unit:            "mg/dL",
	})

	results, _ := repo.FindByClientID(context.Background(), "client-1", 10, 0)
	if results[0].ID == "" {
		t.Error("expected non-empty ID")
	}
	if results[0].CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestInMemoryMeasurementRepository_Add_PreservesExistingID(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	repo.Add(&entities.Measurement{
		ID:              "custom-id",
		ClientID:        "client-1",
		MeasurementType: "ldl",
		Value:           190,
		Unit:            "mg/dL",
	})

	results, _ := repo.FindByClientID(context.Background(), "client-1", 10, 0)
	if results[0].ID != "custom-id" {
		t.Errorf("expected ID custom-id, got %s", results[0].ID)
	}
}

func TestInMemoryMeasurementRepository_Add_PreservesExistingCreatedAt(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	ts := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	repo.Add(&entities.Measurement{
		ClientID:        "client-1",
		MeasurementType: "ldl",
		Value:           190,
		CreatedAt:       ts,
	})

	results, _ := repo.FindByClientID(context.Background(), "client-1", 10, 0)
	if !results[0].CreatedAt.Equal(ts) {
		t.Errorf("expected preserved CreatedAt %v, got %v", ts, results[0].CreatedAt)
	}
}

func TestInMemoryMeasurementRepository_FindByClientID_FiltersCorrectly(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	repo.Add(&entities.Measurement{ClientID: "client-1", MeasurementType: "ldl", Value: 190})
	repo.Add(&entities.Measurement{ClientID: "client-2", MeasurementType: "hdl", Value: 55})
	repo.Add(&entities.Measurement{ClientID: "client-1", MeasurementType: "glucose", Value: 100})

	results, err := repo.FindByClientID(context.Background(), "client-1", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 measurements for client-1, got %d", len(results))
	}
}

func TestInMemoryMeasurementRepository_FindByClientID_Pagination(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	for i := 0; i < 5; i++ {
		repo.Add(&entities.Measurement{ClientID: "client-1", MeasurementType: "ldl", Value: float64(i)})
	}

	results, err := repo.FindByClientID(context.Background(), "client-1", 2, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 with limit=2 offset=1, got %d", len(results))
	}
}

func TestInMemoryMeasurementRepository_FindByClientID_OffsetBeyondLength(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	repo.Add(&entities.Measurement{ClientID: "client-1", MeasurementType: "ldl", Value: 190})

	results, err := repo.FindByClientID(context.Background(), "client-1", 10, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 with offset beyond length, got %d", len(results))
	}
}

func TestInMemoryMeasurementRepository_FindByClientID_DefaultLimit(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	for i := 0; i < 55; i++ {
		repo.Add(&entities.Measurement{ClientID: "client-1", MeasurementType: "ldl", Value: float64(i)})
	}

	// limit=0 should default to 50
	results, err := repo.FindByClientID(context.Background(), "client-1", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 50 {
		t.Errorf("expected default limit of 50, got %d", len(results))
	}
}

func TestInMemoryMeasurementRepository_FindByClientIDAndType_FiltersTypeAndTime(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()

	repo.Add(&entities.Measurement{
		ClientID: "client-1", MeasurementType: "hrv", Value: 55,
		MeasuredAt: now.Add(-2 * time.Hour),
	})
	repo.Add(&entities.Measurement{
		ClientID: "client-1", MeasurementType: "hrv", Value: 50,
		MeasuredAt: now.Add(-48 * time.Hour),
	})
	repo.Add(&entities.Measurement{
		ClientID: "client-1", MeasurementType: "sleep_hours", Value: 7,
		MeasuredAt: now.Add(-2 * time.Hour),
	})
	repo.Add(&entities.Measurement{
		ClientID: "client-2", MeasurementType: "hrv", Value: 60,
		MeasuredAt: now.Add(-2 * time.Hour),
	})

	since := now.Add(-24 * time.Hour)
	results, err := repo.FindByClientIDAndType(context.Background(), "client-1", "hrv", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 HRV measurement within 24h for client-1, got %d", len(results))
	}
	if results[0].Value != 55 {
		t.Errorf("expected value 55, got %f", results[0].Value)
	}
}

func TestInMemoryMeasurementRepository_FindByClientIDAndType_IncludesExactSince(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	since := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	repo.Add(&entities.Measurement{
		ClientID: "client-1", MeasurementType: "hrv", Value: 50,
		MeasuredAt: since, // exactly at the boundary
	})

	results, err := repo.FindByClientIDAndType(context.Background(), "client-1", "hrv", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 measurement at exact boundary, got %d", len(results))
	}
}

func TestInMemoryMeasurementRepository_FindByClientIDAndType_Empty_ReturnsNil(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	results, err := repo.FindByClientIDAndType(context.Background(), "client-1", "hrv", time.Now().Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil, got %v", results)
	}
}

func TestInMemoryMeasurementRepository_FindRecentByArtifactID_FiltersCorrectly(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	repo.Add(&entities.Measurement{
		ClientID: "client-1", MeasurementType: "ldl", Value: 190, ArtifactID: "art-1",
	})
	repo.Add(&entities.Measurement{
		ClientID: "client-1", MeasurementType: "hdl", Value: 55, ArtifactID: "art-1",
	})
	repo.Add(&entities.Measurement{
		ClientID: "client-1", MeasurementType: "glucose", Value: 100, ArtifactID: "art-2",
	})

	results, err := repo.FindRecentByArtifactID(context.Background(), "art-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 measurements for art-1, got %d", len(results))
	}
}

func TestInMemoryMeasurementRepository_FindRecentByArtifactID_NoMatch_ReturnsNil(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	repo.Add(&entities.Measurement{
		ClientID: "client-1", MeasurementType: "ldl", Value: 190, ArtifactID: "art-1",
	})

	results, err := repo.FindRecentByArtifactID(context.Background(), "art-999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil for no match, got %v", results)
	}
}

func TestInMemoryMeasurementRepository_FindRecentByArtifactID_Empty_ReturnsNil(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	results, err := repo.FindRecentByArtifactID(context.Background(), "art-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil on empty repo, got %v", results)
	}
}
