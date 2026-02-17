package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryMeasurementRepository_BulkUpsert_InsertsAll(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	measurements := []entities.WearableMeasurement{
		{
			ClientID:        "client-1",
			Source:          entities.WearableSourceWhoop,
			MeasurementType: entities.MeasurementTypeHRV,
			Value:           45.2,
			Unit:            "ms",
			MeasuredAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ClientID:        "client-1",
			Source:          entities.WearableSourceWhoop,
			MeasurementType: entities.MeasurementTypeSleepDuration,
			Value:           7.5,
			Unit:            "hours",
			MeasuredAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	inserted, err := repo.BulkUpsert(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inserted != 2 {
		t.Errorf("expected 2 inserted, got %d", inserted)
	}
	if repo.Count() != 2 {
		t.Errorf("expected count 2, got %d", repo.Count())
	}
}

func TestInMemoryMeasurementRepository_BulkUpsert_SkipsDuplicates(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	m := entities.WearableMeasurement{
		ClientID:        "client-1",
		Source:          entities.WearableSourceWhoop,
		MeasurementType: entities.MeasurementTypeHRV,
		Value:           45.2,
		Unit:            "ms",
		MeasuredAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// First insert
	inserted, err := repo.BulkUpsert(context.Background(), []entities.WearableMeasurement{m})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inserted != 1 {
		t.Errorf("expected 1 inserted, got %d", inserted)
	}

	// Second insert (duplicate)
	inserted, err = repo.BulkUpsert(context.Background(), []entities.WearableMeasurement{m})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inserted != 0 {
		t.Errorf("expected 0 inserted (duplicate), got %d", inserted)
	}
	if repo.Count() != 1 {
		t.Errorf("expected count still 1, got %d", repo.Count())
	}
}

func TestInMemoryMeasurementRepository_GetAverages_ComputesCorrectly(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	measurements := []entities.WearableMeasurement{
		{
			ClientID:        "client-1",
			Source:          entities.WearableSourceWhoop,
			MeasurementType: entities.MeasurementTypeHRV,
			Value:           40.0,
			Unit:            "ms",
			MeasuredAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ClientID:        "client-1",
			Source:          entities.WearableSourceWhoop,
			MeasurementType: entities.MeasurementTypeHRV,
			Value:           50.0,
			Unit:            "ms",
			MeasuredAt:      time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	_, err := repo.BulkUpsert(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	avg, err := repo.GetAverages(context.Background(), "client-1", entities.WearableSourceWhoop, entities.MeasurementTypeHRV,
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if avg != 45.0 {
		t.Errorf("expected average 45.0, got %f", avg)
	}
}

func TestInMemoryMeasurementRepository_GetAverages_NoData_ReturnsError(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	_, err := repo.GetAverages(context.Background(), "client-1", entities.WearableSourceWhoop, entities.MeasurementTypeHRV,
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("expected error for no data")
	}
}
