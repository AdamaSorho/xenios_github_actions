package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryMeasurementRepo_UpsertBatch_InsertsAll(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	ms := []entities.Measurement{
		{ClientID: "c1", Source: entities.WearableSourceWhoop, MeasurementType: entities.MeasurementTypeHRV, Value: 45, MeasuredAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ClientID: "c1", Source: entities.WearableSourceWhoop, MeasurementType: entities.MeasurementTypeRecovery, Value: 72, MeasuredAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	inserted, err := repo.UpsertBatch(context.Background(), ms)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inserted != 2 {
		t.Errorf("expected 2 inserted, got %d", inserted)
	}
}

func TestInMemoryMeasurementRepo_UpsertBatch_SkipsDuplicates(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	m := entities.Measurement{
		ClientID: "c1", Source: entities.WearableSourceWhoop,
		MeasurementType: entities.MeasurementTypeHRV, Value: 45,
		MeasuredAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// First batch
	inserted, _ := repo.UpsertBatch(context.Background(), []entities.Measurement{m})
	if inserted != 1 {
		t.Errorf("expected 1 inserted, got %d", inserted)
	}

	// Second batch (same data → duplicate)
	inserted, _ = repo.UpsertBatch(context.Background(), []entities.Measurement{m})
	if inserted != 0 {
		t.Errorf("expected 0 inserted (duplicate), got %d", inserted)
	}

	if len(repo.All()) != 1 {
		t.Errorf("expected 1 stored measurement, got %d", len(repo.All()))
	}
}

func TestInMemoryMeasurementRepo_Average_ReturnsCorrectValue(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	base := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	ms := []entities.Measurement{
		{ClientID: "c1", Source: entities.WearableSourceWhoop, MeasurementType: entities.MeasurementTypeHRV, Value: 40, MeasuredAt: base.AddDate(0, 0, -2)},
		{ClientID: "c1", Source: entities.WearableSourceWhoop, MeasurementType: entities.MeasurementTypeHRV, Value: 50, MeasuredAt: base.AddDate(0, 0, -1)},
		{ClientID: "c1", Source: entities.WearableSourceWhoop, MeasurementType: entities.MeasurementTypeHRV, Value: 60, MeasuredAt: base},
	}
	repo.UpsertBatch(context.Background(), ms)

	since := base.AddDate(0, 0, -7)
	avg, err := repo.Average(context.Background(), "c1", entities.WearableSourceWhoop, entities.MeasurementTypeHRV, since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if avg == nil {
		t.Fatal("expected non-nil average")
	}
	if *avg != 50.0 {
		t.Errorf("expected average 50.0, got %f", *avg)
	}
}

func TestInMemoryMeasurementRepo_Average_NoData_ReturnsNil(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	avg, err := repo.Average(context.Background(), "c1", entities.WearableSourceWhoop, entities.MeasurementTypeHRV, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if avg != nil {
		t.Errorf("expected nil for no data, got %f", *avg)
	}
}

func TestInMemoryMeasurementRepo_Average_FiltersBySource(t *testing.T) {
	repo := NewInMemoryMeasurementRepository()
	base := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	ms := []entities.Measurement{
		{ClientID: "c1", Source: entities.WearableSourceWhoop, MeasurementType: entities.MeasurementTypeHRV, Value: 40, MeasuredAt: base},
		{ClientID: "c1", Source: entities.WearableSourceGarmin, MeasurementType: entities.MeasurementTypeHRV, Value: 80, MeasuredAt: base},
	}
	repo.UpsertBatch(context.Background(), ms)

	avg, _ := repo.Average(context.Background(), "c1", entities.WearableSourceWhoop, entities.MeasurementTypeHRV, base.AddDate(0, 0, -7))
	if avg == nil || *avg != 40.0 {
		t.Errorf("expected whoop-only average 40.0, got %v", avg)
	}
}

// ── InMemory Wearable Summary Tests ─────────────────────────────────────────

func TestInMemoryWearableSummaryRepo_Upsert_StoresMetrics(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	metrics := json.RawMessage(`{"avg_hrv_7d": 45.2}`)

	err := repo.Upsert(context.Background(), "c1", entities.WearableSourceWhoop, metrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := repo.Get("c1", entities.WearableSourceWhoop)
	if !ok {
		t.Fatal("expected metrics to be stored")
	}
	if string(got) != string(metrics) {
		t.Errorf("expected %s, got %s", string(metrics), string(got))
	}
}

func TestInMemoryWearableSummaryRepo_Upsert_OverwritesExisting(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()

	repo.Upsert(context.Background(), "c1", entities.WearableSourceWhoop, json.RawMessage(`{"v":1}`))
	repo.Upsert(context.Background(), "c1", entities.WearableSourceWhoop, json.RawMessage(`{"v":2}`))

	got, ok := repo.Get("c1", entities.WearableSourceWhoop)
	if !ok {
		t.Fatal("expected metrics to be stored")
	}
	if string(got) != `{"v":2}` {
		t.Errorf("expected updated metrics, got %s", string(got))
	}
}

func TestInMemoryWearableSummaryRepo_Get_DifferentSources_Independent(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	repo.Upsert(context.Background(), "c1", entities.WearableSourceWhoop, json.RawMessage(`{"src":"whoop"}`))
	repo.Upsert(context.Background(), "c1", entities.WearableSourceGarmin, json.RawMessage(`{"src":"garmin"}`))

	whoop, ok := repo.Get("c1", entities.WearableSourceWhoop)
	if !ok || string(whoop) != `{"src":"whoop"}` {
		t.Errorf("unexpected whoop metrics: %s", string(whoop))
	}

	garmin, ok := repo.Get("c1", entities.WearableSourceGarmin)
	if !ok || string(garmin) != `{"src":"garmin"}` {
		t.Errorf("unexpected garmin metrics: %s", string(garmin))
	}
}
