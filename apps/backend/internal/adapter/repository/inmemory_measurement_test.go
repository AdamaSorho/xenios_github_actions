package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryMeasurementRepository_BatchCreate_StoresMeasurements(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	measurements := []*entities.Measurement{
		{
			ClientID:        "client-1",
			RecordedBy:      "coach-1",
			MeasurementType: entities.MeasurementTypeCalories,
			Value:           2050,
			Unit:            "kcal",
			MeasuredAt:      time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			ClientID:        "client-1",
			RecordedBy:      "coach-1",
			MeasurementType: entities.MeasurementTypeProtein,
			Value:           132,
			Unit:            "g",
			MeasuredAt:      time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	err := repo.BatchCreate(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all := repo.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 measurements, got %d", len(all))
	}
}

func TestInMemoryMeasurementRepository_BatchCreate_AssignsIDs(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	measurements := []*entities.Measurement{
		{
			ClientID:        "client-1",
			MeasurementType: entities.MeasurementTypeCalories,
			Value:           2050,
			MeasuredAt:      time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	err := repo.BatchCreate(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all := repo.GetAll()
	if all[0].ID == "" {
		t.Error("expected measurement to have an assigned ID")
	}
}

func TestInMemoryMeasurementRepository_BatchCreate_SetsCreatedAt(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	before := time.Now()
	measurements := []*entities.Measurement{
		{
			ClientID:        "client-1",
			MeasurementType: entities.MeasurementTypeCalories,
			Value:           2050,
			MeasuredAt:      time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	err := repo.BatchCreate(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all := repo.GetAll()
	if all[0].CreatedAt.Before(before) {
		t.Error("expected CreatedAt to be set to current time")
	}
}

func TestInMemoryMeasurementRepository_FindByClientAndDateRange_FiltersCorrectly(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	measurements := []*entities.Measurement{
		{
			ClientID:        "client-1",
			MeasurementType: entities.MeasurementTypeCalories,
			Value:           2050,
			MeasuredAt:      time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			ClientID:        "client-1",
			MeasurementType: entities.MeasurementTypeCalories,
			Value:           1900,
			MeasuredAt:      time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			ClientID:        "client-2",
			MeasurementType: entities.MeasurementTypeCalories,
			Value:           2200,
			MeasuredAt:      time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	err := repo.BatchCreate(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Filter by client-1 and date range that includes only the first measurement
	from := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)

	result, err := repo.FindByClientAndDateRange(context.Background(), "client-1", from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 measurement, got %d", len(result))
	}
	if result[0].Value != 2050 {
		t.Errorf("expected value 2050, got %.0f", result[0].Value)
	}
}

func TestInMemoryMeasurementRepository_FindByClientAndDateRange_ReturnsEmpty(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	result, err := repo.FindByClientAndDateRange(context.Background(), "client-1", from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(result))
	}
}

func TestInMemoryMeasurementRepository_GetAll_ReturnsCopies(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	measurements := []*entities.Measurement{
		{
			ClientID:        "client-1",
			MeasurementType: entities.MeasurementTypeCalories,
			Value:           2050,
			MeasuredAt:      time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	err := repo.BatchCreate(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all := repo.GetAll()
	all[0].Value = 9999

	// Verify original is unchanged
	all2 := repo.GetAll()
	if all2[0].Value != 2050 {
		t.Errorf("expected original value 2050 to be unchanged, got %.0f", all2[0].Value)
	}
}

func TestInMemoryMeasurementRepository_BatchCreate_PreservesExistingID(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	measurements := []*entities.Measurement{
		{
			ID:              "custom-id",
			ClientID:        "client-1",
			MeasurementType: entities.MeasurementTypeCalories,
			Value:           2050,
			MeasuredAt:      time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	err := repo.BatchCreate(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all := repo.GetAll()
	if all[0].ID != "custom-id" {
		t.Errorf("expected ID 'custom-id', got '%s'", all[0].ID)
	}
}
