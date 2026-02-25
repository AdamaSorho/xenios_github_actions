package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryWearableSummaryRepository_Upsert_Create(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	ctx := context.Background()

	summary := &entities.WearableSummary{
		ClientID:    "client-1",
		Source:      "whoop",
		SummaryDate: "2026-01-15",
		Metrics:     map[string]interface{}{"hrv": 45.0},
	}

	result, err := repo.Upsert(ctx, summary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID == "" {
		t.Error("expected non-empty ID")
	}
	if result.Source != "whoop" {
		t.Errorf("expected source 'whoop', got '%s'", result.Source)
	}
}

func TestInMemoryWearableSummaryRepository_Upsert_Update(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	ctx := context.Background()

	today := time.Now().Format("2006-01-02")

	summary1 := &entities.WearableSummary{
		ClientID:    "client-1",
		Source:      "whoop",
		SummaryDate: today,
		Metrics:     map[string]interface{}{"hrv": 45.0},
	}
	first, _ := repo.Upsert(ctx, summary1)

	summary2 := &entities.WearableSummary{
		ClientID:    "client-1",
		Source:      "whoop",
		SummaryDate: today,
		Metrics:     map[string]interface{}{"hrv": 50.0},
	}
	second, err := repo.Upsert(ctx, summary2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if second.ID != first.ID {
		t.Errorf("expected same ID on update, got %s vs %s", first.ID, second.ID)
	}

	// Verify only 1 record exists
	results, _ := repo.FindByClientID(ctx, "client-1", 7)
	if len(results) != 1 {
		t.Errorf("expected 1 record after upsert, got %d", len(results))
	}
}

func TestInMemoryWearableSummaryRepository_FindByClientID(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	ctx := context.Background()

	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	oldDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

	_, _ = repo.Upsert(ctx, &entities.WearableSummary{ClientID: "client-1", Source: "whoop", SummaryDate: today, Metrics: map[string]interface{}{"hrv": 45.0}})
	_, _ = repo.Upsert(ctx, &entities.WearableSummary{ClientID: "client-1", Source: "whoop", SummaryDate: yesterday, Metrics: map[string]interface{}{"hrv": 50.0}})
	_, _ = repo.Upsert(ctx, &entities.WearableSummary{ClientID: "client-1", Source: "whoop", SummaryDate: oldDate, Metrics: map[string]interface{}{"hrv": 40.0}})

	result, err := repo.FindByClientID(ctx, "client-1", 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 summaries within 7 days, got %d", len(result))
	}
}

func TestInMemoryWearableSummaryRepository_FindByClientID_Empty(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	ctx := context.Background()

	result, err := repo.FindByClientID(ctx, "client-1", 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil && len(result) != 0 {
		t.Errorf("expected empty result, got %d", len(result))
	}
}

func TestInMemoryWearableSummaryRepository_FindByClientID_FiltersByClient(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	ctx := context.Background()

	today := time.Now().Format("2006-01-02")
	_, _ = repo.Upsert(ctx, &entities.WearableSummary{ClientID: "client-1", Source: "whoop", SummaryDate: today, Metrics: map[string]interface{}{"hrv": 45.0}})
	_, _ = repo.Upsert(ctx, &entities.WearableSummary{ClientID: "client-2", Source: "whoop", SummaryDate: today, Metrics: map[string]interface{}{"hrv": 50.0}})

	result, err := repo.FindByClientID(ctx, "client-1", 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 summary for client-1, got %d", len(result))
	}
}
