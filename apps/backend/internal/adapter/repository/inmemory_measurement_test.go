package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/repository"
)

func TestInMemoryMeasurementRepository_BatchCreate_StoresMeasurements(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	measurements := []repository.Measurement{
		{ClientID: "c1", RecordedBy: "coach1", MeasurementType: "calories", Value: 2050, Unit: "kcal", MeasuredAt: time.Now()},
		{ClientID: "c1", RecordedBy: "coach1", MeasurementType: "protein", Value: 132, Unit: "g", MeasuredAt: time.Now()},
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

	measurements := []repository.Measurement{
		{ClientID: "c1", MeasurementType: "calories", Value: 2000, Unit: "kcal", MeasuredAt: time.Now()},
	}

	err := repo.BatchCreate(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all := repo.GetAll()
	if all[0].ID == "" {
		t.Error("expected measurement to have an ID")
	}
}

func TestInMemoryMeasurementRepository_BatchCreate_EmptySlice_NoError(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	err := repo.BatchCreate(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all := repo.GetAll()
	if len(all) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(all))
	}
}
