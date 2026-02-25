package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryMeasurementRepository_Create(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	ctx := context.Background()

	m := &entities.Measurement{
		ClientID:   "client-1",
		Type:       "weight",
		Value:      185.4,
		Unit:       "lbs",
		MeasuredAt: time.Now(),
		RecordedBy: "coach-1",
	}

	result, err := repo.Create(ctx, m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID == "" {
		t.Error("expected non-empty ID")
	}
	if result.Value != 185.4 {
		t.Errorf("expected value 185.4, got %f", result.Value)
	}
}

func TestInMemoryMeasurementRepository_FindByClientID_Basic(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	ctx := context.Background()

	now := time.Now()
	_, _ = repo.Create(ctx, &entities.Measurement{ClientID: "client-1", Type: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now, RecordedBy: "coach-1"})
	_, _ = repo.Create(ctx, &entities.Measurement{ClientID: "client-1", Type: "body_fat_pct", Value: 22.3, Unit: "%", MeasuredAt: now, RecordedBy: "coach-1"})
	_, _ = repo.Create(ctx, &entities.Measurement{ClientID: "client-2", Type: "weight", Value: 200, Unit: "lbs", MeasuredAt: now, RecordedBy: "coach-1"})

	result, total, err := repo.FindByClientID(ctx, entities.MeasurementFilter{
		ClientID: "client-1",
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 measurements, got %d", len(result))
	}
}

func TestInMemoryMeasurementRepository_FindByClientID_TypeFilter(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	ctx := context.Background()

	now := time.Now()
	_, _ = repo.Create(ctx, &entities.Measurement{ClientID: "client-1", Type: "weight", Value: 185, Unit: "lbs", MeasuredAt: now, RecordedBy: "coach-1"})
	_, _ = repo.Create(ctx, &entities.Measurement{ClientID: "client-1", Type: "body_fat_pct", Value: 22, Unit: "%", MeasuredAt: now, RecordedBy: "coach-1"})

	result, total, err := repo.FindByClientID(ctx, entities.MeasurementFilter{
		ClientID: "client-1",
		Type:     "weight",
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 measurement, got %d", len(result))
	}
	if result[0].Type != "weight" {
		t.Errorf("expected type 'weight', got '%s'", result[0].Type)
	}
}

func TestInMemoryMeasurementRepository_FindByClientID_DateRange(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	ctx := context.Background()

	jan1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	jan15 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	feb1 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	feb15 := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)

	_, _ = repo.Create(ctx, &entities.Measurement{ClientID: "client-1", Type: "weight", Value: 185, Unit: "lbs", MeasuredAt: jan1, RecordedBy: "coach-1"})
	_, _ = repo.Create(ctx, &entities.Measurement{ClientID: "client-1", Type: "weight", Value: 184, Unit: "lbs", MeasuredAt: jan15, RecordedBy: "coach-1"})
	_, _ = repo.Create(ctx, &entities.Measurement{ClientID: "client-1", Type: "weight", Value: 183, Unit: "lbs", MeasuredAt: feb15, RecordedBy: "coach-1"})

	from := jan1
	to := feb1
	result, total, err := repo.FindByClientID(ctx, entities.MeasurementFilter{
		ClientID: "client-1",
		From:     &from,
		To:       &to,
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 measurements, got %d", len(result))
	}
}

func TestInMemoryMeasurementRepository_FindByClientID_Pagination(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	ctx := context.Background()

	now := time.Now()
	for i := 0; i < 5; i++ {
		_, _ = repo.Create(ctx, &entities.Measurement{
			ClientID: "client-1", Type: "weight",
			Value: float64(180 + i), Unit: "lbs",
			MeasuredAt: now.Add(time.Duration(i) * time.Hour),
			RecordedBy: "coach-1",
		})
	}

	result, total, err := repo.FindByClientID(ctx, entities.MeasurementFilter{
		ClientID: "client-1",
		Limit:    2,
		Offset:   2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 measurements, got %d", len(result))
	}
}

func TestInMemoryMeasurementRepository_FindByClientID_EmptyResult(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	ctx := context.Background()

	result, total, err := repo.FindByClientID(ctx, entities.MeasurementFilter{
		ClientID: "client-1",
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(result))
	}
}

func TestInMemoryMeasurementRepository_FindLatestByClientID(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	ctx := context.Background()

	now := time.Now()
	earlier := now.Add(-24 * time.Hour)

	_, _ = repo.Create(ctx, &entities.Measurement{ClientID: "client-1", Type: "weight", Value: 185, Unit: "lbs", MeasuredAt: earlier, RecordedBy: "coach-1"})
	_, _ = repo.Create(ctx, &entities.Measurement{ClientID: "client-1", Type: "weight", Value: 184, Unit: "lbs", MeasuredAt: now, RecordedBy: "coach-1"})
	_, _ = repo.Create(ctx, &entities.Measurement{ClientID: "client-1", Type: "body_fat_pct", Value: 22, Unit: "%", MeasuredAt: earlier, RecordedBy: "coach-1"})

	result, err := repo.FindLatestByClientID(ctx, "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 measurements (one per type), got %d", len(result))
	}

	// Find the weight measurement
	for _, m := range result {
		if m.Type == "weight" && m.Value != 184 {
			t.Errorf("expected latest weight 184, got %f", m.Value)
		}
	}
}

func TestInMemoryMeasurementRepository_FindByType(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	ctx := context.Background()

	now := time.Now()
	_, _ = repo.Create(ctx, &entities.Measurement{ClientID: "client-1", Type: "weight", Value: 185, Unit: "lbs", MeasuredAt: now, RecordedBy: "coach-1"})
	_, _ = repo.Create(ctx, &entities.Measurement{ClientID: "client-1", Type: "body_fat_pct", Value: 22, Unit: "%", MeasuredAt: now, RecordedBy: "coach-1"})

	result, err := repo.FindByType(ctx, "client-1", "weight")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 measurement, got %d", len(result))
	}
}
