package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryMeasurementRepository_CreateBatch_StoresMeasurements(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	measurements := []*entities.Measurement{
		{ClientID: "c1", RecordedBy: "coach-1", MeasurementType: "calories", Value: 2050, Unit: "kcal", MeasuredAt: time.Now()},
		{ClientID: "c1", RecordedBy: "coach-1", MeasurementType: "protein", Value: 150, Unit: "g", MeasuredAt: time.Now()},
	}

	err := repo.CreateBatch(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all := repo.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 measurements, got %d", len(all))
	}
}

func TestInMemoryMeasurementRepository_CreateBatch_AssignsIDs(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	measurements := []*entities.Measurement{
		{ClientID: "c1", MeasurementType: "calories", Value: 2050, Unit: "kcal", MeasuredAt: time.Now()},
	}

	err := repo.CreateBatch(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all := repo.GetAll()
	if all[0].ID == "" {
		t.Error("expected ID to be assigned")
	}
}

func TestInMemoryMeasurementRepository_CreateBatch_Empty_NoError(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	err := repo.CreateBatch(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all := repo.GetAll()
	if len(all) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(all))
	}
}

func TestInMemoryMeasurementRepository_CreateBatch_SetsCreatedAt(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()

	measurements := []*entities.Measurement{
		{ClientID: "c1", MeasurementType: "calories", Value: 2050, Unit: "kcal", MeasuredAt: time.Now()},
	}

	err := repo.CreateBatch(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all := repo.GetAll()
	if all[0].CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}
