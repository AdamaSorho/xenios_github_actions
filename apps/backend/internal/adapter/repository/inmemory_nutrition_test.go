package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryNutritionRepository_SaveRecords_StoresRecords(t *testing.T) {
	repo := NewInMemoryNutritionRepository()
	records := []*entities.NutritionRecord{
		{ClientID: "c1", CoachID: "coach1", ArtifactID: "a1", MetricType: entities.NutritionMetricCalories, Value: 2000, Unit: "kcal", RecordDate: time.Now()},
		{ClientID: "c1", CoachID: "coach1", ArtifactID: "a1", MetricType: entities.NutritionMetricProtein, Value: 150, Unit: "g", RecordDate: time.Now()},
	}

	err := repo.SaveRecords(context.Background(), records)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored := repo.GetRecords()
	if len(stored) != 2 {
		t.Errorf("expected 2 records, got %d", len(stored))
	}
}

func TestInMemoryNutritionRepository_UpsertSummary_CreatesSummary(t *testing.T) {
	repo := NewInMemoryNutritionRepository()
	summary := &entities.NutritionSummary{
		ClientID:      "c1",
		ArtifactID:    "a1",
		TotalDays:     7,
		AvgCalories7d: 2000,
	}

	err := repo.UpsertSummary(context.Background(), summary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := repo.GetSummaryByClientID(context.Background(), "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected summary, got nil")
	}
	if result.AvgCalories7d != 2000 {
		t.Errorf("expected avg_calories_7d 2000, got %f", result.AvgCalories7d)
	}
}

func TestInMemoryNutritionRepository_UpsertSummary_UpdatesExisting(t *testing.T) {
	repo := NewInMemoryNutritionRepository()
	summary1 := &entities.NutritionSummary{
		ClientID:      "c1",
		ArtifactID:    "a1",
		TotalDays:     7,
		AvgCalories7d: 2000,
	}
	_ = repo.UpsertSummary(context.Background(), summary1)

	summary2 := &entities.NutritionSummary{
		ClientID:      "c1",
		ArtifactID:    "a1",
		TotalDays:     14,
		AvgCalories7d: 2200,
	}
	_ = repo.UpsertSummary(context.Background(), summary2)

	result, _ := repo.GetSummaryByClientID(context.Background(), "c1")
	if result.AvgCalories7d != 2200 {
		t.Errorf("expected updated avg_calories_7d 2200, got %f", result.AvgCalories7d)
	}
}

func TestInMemoryNutritionRepository_GetSummaryByClientID_NotFound_ReturnsNil(t *testing.T) {
	repo := NewInMemoryNutritionRepository()
	result, err := repo.GetSummaryByClientID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil, got non-nil summary")
	}
}
