package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryMeasurementRepository_FindRecentByClientID_ReturnsRecent(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()

	repo.Add(&entities.Measurement{
		ID: "m-1", ClientID: "client-1", MeasurementType: "weight",
		Value: 180, Unit: "lbs", MeasuredAt: now,
	})
	repo.Add(&entities.Measurement{
		ID: "m-2", ClientID: "client-1", MeasurementType: "weight",
		Value: 175, Unit: "lbs", MeasuredAt: now.AddDate(0, 0, -60), // Too old
	})

	results, err := repo.FindRecentByClientID(context.Background(), "client-1", now.AddDate(0, 0, -30))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestInMemoryMeasurementRepository_FindRecentByClientID_FiltersByClient(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()

	repo.Add(&entities.Measurement{
		ID: "m-1", ClientID: "client-1", MeasurementType: "weight",
		Value: 180, Unit: "lbs", MeasuredAt: now,
	})
	repo.Add(&entities.Measurement{
		ID: "m-2", ClientID: "client-2", MeasurementType: "weight",
		Value: 200, Unit: "lbs", MeasuredAt: now,
	})

	results, err := repo.FindRecentByClientID(context.Background(), "client-1", now.AddDate(0, 0, -30))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestInMemoryMeasurementRepository_FindByClientIDAndType_FiltersCorrectly(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()

	repo.Add(&entities.Measurement{
		ID: "m-1", ClientID: "client-1", MeasurementType: "weight",
		Value: 180, Unit: "lbs", MeasuredAt: now,
	})
	repo.Add(&entities.Measurement{
		ID: "m-2", ClientID: "client-1", MeasurementType: "body_fat_percentage",
		Value: 20, Unit: "%", MeasuredAt: now,
	})

	results, err := repo.FindByClientIDAndType(context.Background(), "client-1", "weight", now.AddDate(0, 0, -30))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].MeasurementType != "weight" {
		t.Errorf("expected type 'weight', got %q", results[0].MeasurementType)
	}
}

func TestInMemoryMeasurementRepository_FindRecentByClientID_Empty_ReturnsNil(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	results, err := repo.FindRecentByClientID(context.Background(), "client-1", time.Now().AddDate(0, 0, -30))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil, got %v", results)
	}
}
