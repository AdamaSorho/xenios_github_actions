package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryWearableSummaryRepository_FindByClientIDAndDateRange_ReturnsMatching(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	now := time.Now()

	repo.Add(&entities.WearableSummary{
		ID: "ws-1", ClientID: "client-1", Source: "whoop",
		SummaryDate: now.AddDate(0, 0, -3),
		Metrics:     map[string]interface{}{"hrv": float64(55)},
		SyncedAt:    now, CreatedAt: now,
	})
	repo.Add(&entities.WearableSummary{
		ID: "ws-2", ClientID: "client-1", Source: "whoop",
		SummaryDate: now.AddDate(0, 0, -20), // Outside range
		Metrics:     map[string]interface{}{"hrv": float64(60)},
		SyncedAt:    now, CreatedAt: now,
	})

	results, err := repo.FindByClientIDAndDateRange(context.Background(), "client-1", now.AddDate(0, 0, -7), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestInMemoryWearableSummaryRepository_FindByClientIDAndDateRange_FiltersByClient(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	now := time.Now()

	repo.Add(&entities.WearableSummary{
		ID: "ws-1", ClientID: "client-1", Source: "whoop",
		SummaryDate: now.AddDate(0, 0, -3),
		Metrics:     map[string]interface{}{"hrv": float64(55)},
		SyncedAt:    now, CreatedAt: now,
	})
	repo.Add(&entities.WearableSummary{
		ID: "ws-2", ClientID: "client-2", Source: "whoop",
		SummaryDate: now.AddDate(0, 0, -3),
		Metrics:     map[string]interface{}{"hrv": float64(60)},
		SyncedAt:    now, CreatedAt: now,
	})

	results, err := repo.FindByClientIDAndDateRange(context.Background(), "client-1", now.AddDate(0, 0, -7), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestInMemoryWearableSummaryRepository_FindByClientIDAndDateRange_Empty_ReturnsNil(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	results, err := repo.FindByClientIDAndDateRange(context.Background(), "client-1", time.Now().AddDate(0, 0, -7), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil, got %v", results)
	}
}

func TestInMemoryWearableSummaryRepository_FindByClientIDAndDateRange_CopiesMetrics(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	now := time.Now()

	repo.Add(&entities.WearableSummary{
		ID: "ws-1", ClientID: "client-1", Source: "whoop",
		SummaryDate: now,
		Metrics:     map[string]interface{}{"hrv": float64(55)},
		SyncedAt:    now, CreatedAt: now,
	})

	results, _ := repo.FindByClientIDAndDateRange(context.Background(), "client-1", now.AddDate(0, 0, -1), now.AddDate(0, 0, 1))
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	// Modify returned metrics — should not affect stored data
	results[0].Metrics["hrv"] = float64(999)

	results2, _ := repo.FindByClientIDAndDateRange(context.Background(), "client-1", now.AddDate(0, 0, -1), now.AddDate(0, 0, 1))
	hrv, ok := results2[0].GetMetricFloat64("hrv")
	if !ok || hrv != 55 {
		t.Errorf("expected original HRV 55, got %f", hrv)
	}
}
