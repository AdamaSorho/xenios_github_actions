package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryWearableSummaryRepository_Upsert_NewSummary_Stores(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	summary := &entities.WearableSummary{
		ClientID:    "client-1",
		Source:      entities.WearableSourceWhoop,
		SummaryDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Metrics: map[string]interface{}{
			"avg_hrv_7d": 45.0,
		},
		SyncedAt: time.Now(),
	}

	err := repo.Upsert(context.Background(), summary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	summaries := repo.GetSummaries()
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].ClientID != "client-1" {
		t.Errorf("expected client_id %q, got %q", "client-1", summaries[0].ClientID)
	}
}

func TestInMemoryWearableSummaryRepository_Upsert_DuplicateKey_Updates(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	s1 := &entities.WearableSummary{
		ClientID:    "client-1",
		Source:      entities.WearableSourceWhoop,
		SummaryDate: date,
		Metrics:     map[string]interface{}{"avg_hrv_7d": 40.0},
		SyncedAt:    time.Now(),
	}
	_ = repo.Upsert(context.Background(), s1)

	s2 := &entities.WearableSummary{
		ClientID:    "client-1",
		Source:      entities.WearableSourceWhoop,
		SummaryDate: date,
		Metrics:     map[string]interface{}{"avg_hrv_7d": 50.0},
		SyncedAt:    time.Now(),
	}
	_ = repo.Upsert(context.Background(), s2)

	summaries := repo.GetSummaries()
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary after upsert, got %d", len(summaries))
	}
	if summaries[0].Metrics["avg_hrv_7d"] != 50.0 {
		t.Errorf("expected updated avg_hrv_7d 50.0, got %v", summaries[0].Metrics["avg_hrv_7d"])
	}
}

func TestInMemoryWearableSummaryRepository_Upsert_DifferentSources_StoresSeparately(t *testing.T) {
	repo := NewInMemoryWearableSummaryRepository()
	date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	s1 := &entities.WearableSummary{
		ClientID:    "client-1",
		Source:      entities.WearableSourceWhoop,
		SummaryDate: date,
		Metrics:     map[string]interface{}{"avg_hrv_7d": 40.0},
	}
	_ = repo.Upsert(context.Background(), s1)

	s2 := &entities.WearableSummary{
		ClientID:    "client-1",
		Source:      entities.WearableSourceGarmin,
		SummaryDate: date,
		Metrics:     map[string]interface{}{"avg_hrv_7d": 50.0},
	}
	_ = repo.Upsert(context.Background(), s2)

	summaries := repo.GetSummaries()
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries for different sources, got %d", len(summaries))
	}
}
