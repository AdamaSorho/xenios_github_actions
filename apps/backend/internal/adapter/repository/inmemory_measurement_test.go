package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryMeasurementRepo_Create_StoresRecord(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	m := &entities.Measurement{
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "weight",
		Value:           185.4,
		Unit:            "lbs",
		MeasuredAt:      time.Now(),
	}

	created, err := repo.Create(context.Background(), m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == "" {
		t.Error("expected non-empty ID")
	}
	if created.Value != 185.4 {
		t.Errorf("expected value 185.4, got %f", created.Value)
	}
}

func TestInMemoryMeasurementRepo_FindByClientID_ReturnsFiltered(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()

	repo.Create(context.Background(), &entities.Measurement{ClientID: "client-1", MeasurementType: "weight", Value: 185, Unit: "lbs", MeasuredAt: now})
	repo.Create(context.Background(), &entities.Measurement{ClientID: "client-1", MeasurementType: "body_fat_pct", Value: 22, Unit: "%", MeasuredAt: now})
	repo.Create(context.Background(), &entities.Measurement{ClientID: "client-2", MeasurementType: "weight", Value: 200, Unit: "lbs", MeasuredAt: now})

	filter := entities.MeasurementFilter{ClientID: "client-1", Limit: 20}
	results, total, err := repo.FindByClientID(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestInMemoryMeasurementRepo_FindByClientID_FilterByType(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()

	repo.Create(context.Background(), &entities.Measurement{ClientID: "client-1", MeasurementType: "weight", Value: 185, Unit: "lbs", MeasuredAt: now})
	repo.Create(context.Background(), &entities.Measurement{ClientID: "client-1", MeasurementType: "body_fat_pct", Value: 22, Unit: "%", MeasuredAt: now})

	filter := entities.MeasurementFilter{ClientID: "client-1", MeasurementType: "weight", Limit: 20}
	results, total, err := repo.FindByClientID(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].MeasurementType != "weight" {
		t.Errorf("expected type weight, got %s", results[0].MeasurementType)
	}
}

func TestInMemoryMeasurementRepo_FindByClientID_FilterByDateRange(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	jan15 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	feb15 := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	mar15 := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	repo.Create(context.Background(), &entities.Measurement{ClientID: "client-1", MeasurementType: "weight", Value: 185, Unit: "lbs", MeasuredAt: jan15})
	repo.Create(context.Background(), &entities.Measurement{ClientID: "client-1", MeasurementType: "weight", Value: 183, Unit: "lbs", MeasuredAt: feb15})
	repo.Create(context.Background(), &entities.Measurement{ClientID: "client-1", MeasurementType: "weight", Value: 181, Unit: "lbs", MeasuredAt: mar15})

	from := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	filter := entities.MeasurementFilter{ClientID: "client-1", From: &from, To: &to, Limit: 20}
	results, total, err := repo.FindByClientID(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestInMemoryMeasurementRepo_FindByClientID_Pagination(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()

	for i := 0; i < 5; i++ {
		repo.Create(context.Background(), &entities.Measurement{
			ClientID: "client-1", MeasurementType: "weight", Value: float64(180 + i), Unit: "lbs",
			MeasuredAt: now.Add(time.Duration(i) * time.Hour),
		})
	}

	filter := entities.MeasurementFilter{ClientID: "client-1", Limit: 2, Offset: 2}
	results, total, err := repo.FindByClientID(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestInMemoryMeasurementRepo_FindByClientID_OffsetBeyondTotal(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	now := time.Now()
	repo.Create(context.Background(), &entities.Measurement{ClientID: "client-1", MeasurementType: "weight", Value: 185, Unit: "lbs", MeasuredAt: now})

	filter := entities.MeasurementFilter{ClientID: "client-1", Limit: 20, Offset: 100}
	results, total, err := repo.FindByClientID(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestInMemoryMeasurementRepo_FindLatestByClientID_ReturnsLatestPerType(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	older := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	repo.Create(context.Background(), &entities.Measurement{ClientID: "client-1", MeasurementType: "weight", Value: 185, Unit: "lbs", MeasuredAt: older})
	repo.Create(context.Background(), &entities.Measurement{ClientID: "client-1", MeasurementType: "weight", Value: 183, Unit: "lbs", MeasuredAt: newer})
	repo.Create(context.Background(), &entities.Measurement{ClientID: "client-1", MeasurementType: "body_fat_pct", Value: 22, Unit: "%", MeasuredAt: older})

	results, err := repo.FindLatestByClientID(context.Background(), "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results (one per type), got %d", len(results))
	}

	for _, m := range results {
		if m.MeasurementType == "weight" && m.Value != 183 {
			t.Errorf("expected latest weight 183, got %f", m.Value)
		}
	}
}

func TestInMemoryMeasurementRepo_FindLatestByClientID_EmptyForUnknownClient(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	results, err := repo.FindLatestByClientID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
