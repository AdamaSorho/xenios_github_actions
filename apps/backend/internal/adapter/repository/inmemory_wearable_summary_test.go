package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryWearableSummaryRepository_Upsert_CreatesNew(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()

	summary := &entities.WearableSummary{
		ClientID:    "client-1",
		Source:      entities.WearableSourceWhoop,
		SummaryDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Metrics: map[string]interface{}{
			"avg_hrv_7d": 45.2,
		},
	}

	err := repo.Upsert(context.Background(), summary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.Count() != 1 {
		t.Errorf("expected count 1, got %d", repo.Count())
	}
}

func TestInMemoryWearableSummaryRepository_Upsert_UpdatesExisting(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()

	summary := &entities.WearableSummary{
		ClientID:    "client-1",
		Source:      entities.WearableSourceWhoop,
		SummaryDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Metrics: map[string]interface{}{
			"avg_hrv_7d": 45.2,
		},
	}

	_ = repo.Upsert(context.Background(), summary)

	// Update
	summary.Metrics["avg_hrv_7d"] = 48.0
	_ = repo.Upsert(context.Background(), summary)

	if repo.Count() != 1 {
		t.Errorf("expected count 1 after upsert, got %d", repo.Count())
	}

	all := repo.GetAll()
	if len(all) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(all))
	}
	if all[0].Metrics["avg_hrv_7d"] != 48.0 {
		t.Errorf("expected updated avg_hrv_7d 48.0, got %v", all[0].Metrics["avg_hrv_7d"])
	}
}
