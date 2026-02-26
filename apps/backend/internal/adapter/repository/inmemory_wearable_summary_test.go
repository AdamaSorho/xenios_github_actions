package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryWearableRepo_Upsert_Creates(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	ws := &entities.WearableSummary{
		ClientID:    "client-1",
		Source:      "whoop",
		SummaryDate: "2026-01-15",
		Metrics:     map[string]interface{}{"hrv": 45.0},
	}

	created, err := repo.Upsert(context.Background(), ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == "" {
		t.Error("expected non-empty ID")
	}
	if created.Source != "whoop" {
		t.Errorf("expected source whoop, got %s", created.Source)
	}
}

func TestInMemoryWearableRepo_Upsert_Updates(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	ws := &entities.WearableSummary{
		ClientID:    "client-1",
		Source:      "whoop",
		SummaryDate: "2026-01-15",
		Metrics:     map[string]interface{}{"hrv": 45.0},
	}

	first, _ := repo.Upsert(context.Background(), ws)

	ws.Metrics = map[string]interface{}{"hrv": 50.0}
	updated, err := repo.Upsert(context.Background(), ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ID != first.ID {
		t.Errorf("expected same ID on upsert, got %s vs %s", first.ID, updated.ID)
	}

	results, _ := repo.FindByClientID(context.Background(), "client-1", 20, 0)
	if len(results) != 1 {
		t.Errorf("expected 1 record after upsert, got %d", len(results))
	}
}

func TestInMemoryWearableRepo_FindByClientID_ReturnsFiltered(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()

	repo.Upsert(context.Background(), &entities.WearableSummary{ClientID: "client-1", Source: "whoop", SummaryDate: "2026-01-15", Metrics: map[string]interface{}{}})
	repo.Upsert(context.Background(), &entities.WearableSummary{ClientID: "client-1", Source: "whoop", SummaryDate: "2026-01-14", Metrics: map[string]interface{}{}})
	repo.Upsert(context.Background(), &entities.WearableSummary{ClientID: "client-2", Source: "whoop", SummaryDate: "2026-01-15", Metrics: map[string]interface{}{}})

	results, err := repo.FindByClientID(context.Background(), "client-1", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestInMemoryWearableRepo_FindByClientID_OrderedByDateDesc(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()

	repo.Upsert(context.Background(), &entities.WearableSummary{ClientID: "client-1", Source: "whoop", SummaryDate: "2026-01-10", Metrics: map[string]interface{}{}})
	repo.Upsert(context.Background(), &entities.WearableSummary{ClientID: "client-1", Source: "whoop", SummaryDate: "2026-01-15", Metrics: map[string]interface{}{}})
	repo.Upsert(context.Background(), &entities.WearableSummary{ClientID: "client-1", Source: "whoop", SummaryDate: "2026-01-12", Metrics: map[string]interface{}{}})

	results, err := repo.FindByClientID(context.Background(), "client-1", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].SummaryDate != "2026-01-15" {
		t.Errorf("expected first result date 2026-01-15, got %s", results[0].SummaryDate)
	}
	if results[2].SummaryDate != "2026-01-10" {
		t.Errorf("expected last result date 2026-01-10, got %s", results[2].SummaryDate)
	}
}

func TestInMemoryWearableRepo_FindByClientID_Pagination(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()

	for i := 0; i < 5; i++ {
		repo.Upsert(context.Background(), &entities.WearableSummary{
			ClientID: "client-1", Source: "whoop",
			SummaryDate: "2026-01-1" + string(rune('0'+i)),
			Metrics:     map[string]interface{}{},
		})
	}

	results, err := repo.FindByClientID(context.Background(), "client-1", 2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestInMemoryWearableRepo_FindByClientID_OffsetBeyondTotal(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	repo.Upsert(context.Background(), &entities.WearableSummary{ClientID: "client-1", Source: "whoop", SummaryDate: "2026-01-15", Metrics: map[string]interface{}{}})

	results, err := repo.FindByClientID(context.Background(), "client-1", 20, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestInMemoryWearableRepo_FindByClientID_EmptyForUnknownClient(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	results, err := repo.FindByClientID(context.Background(), "nonexistent", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
