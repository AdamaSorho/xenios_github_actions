package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryMeasurementRepository_UpsertBatch_NewMeasurements_InsertsAll(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	measurements := []entities.Measurement{
		{
			ClientID:        "client-1",
			MeasurementType: entities.MeasurementTypeHRV,
			Value:           45.2,
			MeasuredAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Source:          entities.WearableSourceWhoop,
		},
		{
			ClientID:        "client-1",
			MeasurementType: entities.MeasurementTypeSleepDuration,
			Value:           7.5,
			MeasuredAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Source:          entities.WearableSourceWhoop,
		},
	}

	inserted, err := repo.UpsertBatch(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inserted != 2 {
		t.Errorf("expected 2 inserted, got %d", inserted)
	}
}

func TestInMemoryMeasurementRepository_UpsertBatch_Duplicates_SkipsExisting(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	m := entities.Measurement{
		ClientID:        "client-1",
		MeasurementType: entities.MeasurementTypeHRV,
		Value:           45.2,
		MeasuredAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Source:          entities.WearableSourceWhoop,
	}

	// First insert
	inserted1, _ := repo.UpsertBatch(context.Background(), []entities.Measurement{m})
	if inserted1 != 1 {
		t.Errorf("expected 1 on first insert, got %d", inserted1)
	}

	// Second insert (duplicate)
	inserted2, _ := repo.UpsertBatch(context.Background(), []entities.Measurement{m})
	if inserted2 != 0 {
		t.Errorf("expected 0 on duplicate insert, got %d", inserted2)
	}

	// Total should be 1
	all := repo.GetMeasurements()
	if len(all) != 1 {
		t.Errorf("expected 1 total, got %d", len(all))
	}
}

func TestInMemoryMeasurementRepository_Average_WithData_ReturnsCorrectAverage(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()
	measurements := []entities.Measurement{
		{
			ClientID:        "client-1",
			MeasurementType: entities.MeasurementTypeHRV,
			Value:           40.0,
			MeasuredAt:      now.AddDate(0, 0, -1),
			Source:          entities.WearableSourceWhoop,
		},
		{
			ClientID:        "client-1",
			MeasurementType: entities.MeasurementTypeHRV,
			Value:           50.0,
			MeasuredAt:      now.AddDate(0, 0, -2),
			Source:          entities.WearableSourceWhoop,
		},
	}

	_, _ = repo.UpsertBatch(context.Background(), measurements)

	avg, err := repo.Average(context.Background(), "client-1", entities.MeasurementTypeHRV, entities.WearableSourceWhoop, now.AddDate(0, 0, -7))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if avg != 45.0 {
		t.Errorf("expected average 45.0, got %f", avg)
	}
}

func TestInMemoryMeasurementRepository_Average_NoData_ReturnsZero(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	avg, err := repo.Average(context.Background(), "client-1", entities.MeasurementTypeHRV, entities.WearableSourceWhoop, time.Now().AddDate(0, 0, -7))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if avg != 0 {
		t.Errorf("expected 0, got %f", avg)
	}
}

func TestInMemoryMeasurementRepository_Average_FiltersBySource(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()
	measurements := []entities.Measurement{
		{
			ClientID:        "client-1",
			MeasurementType: entities.MeasurementTypeHRV,
			Value:           40.0,
			MeasuredAt:      now.AddDate(0, 0, -1),
			Source:          entities.WearableSourceWhoop,
		},
		{
			ClientID:        "client-1",
			MeasurementType: entities.MeasurementTypeHRV,
			Value:           60.0,
			MeasuredAt:      now.AddDate(0, 0, -1),
			Source:          entities.WearableSourceGarmin,
		},
	}

	_, _ = repo.UpsertBatch(context.Background(), measurements)

	avg, err := repo.Average(context.Background(), "client-1", entities.MeasurementTypeHRV, entities.WearableSourceWhoop, now.AddDate(0, 0, -7))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if avg != 40.0 {
		t.Errorf("expected average 40.0 (WHOOP only), got %f", avg)
	}
}

func TestInMemoryMeasurementRepository_Average_FiltersByTimeWindow(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()
	measurements := []entities.Measurement{
		{
			ClientID:        "client-1",
			MeasurementType: entities.MeasurementTypeHRV,
			Value:           40.0,
			MeasuredAt:      now.AddDate(0, 0, -3),
			Source:          entities.WearableSourceWhoop,
		},
		{
			ClientID:        "client-1",
			MeasurementType: entities.MeasurementTypeHRV,
			Value:           60.0,
			MeasuredAt:      now.AddDate(0, 0, -10),
			Source:          entities.WearableSourceWhoop,
		},
	}

	_, _ = repo.UpsertBatch(context.Background(), measurements)

	// 7-day window should only include the first measurement
	avg, err := repo.Average(context.Background(), "client-1", entities.MeasurementTypeHRV, entities.WearableSourceWhoop, now.AddDate(0, 0, -7))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if avg != 40.0 {
		t.Errorf("expected average 40.0 (7-day window), got %f", avg)
	}
}
