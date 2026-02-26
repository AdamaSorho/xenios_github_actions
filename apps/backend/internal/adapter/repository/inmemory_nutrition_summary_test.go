package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryNutritionSummaryRepository_Upsert_CreatesNew(t *testing.T) {
	repo := NewInMemoryNutritionSummaryRepository()

	summary := &entities.NutritionSummary{
		ClientID:      "c1",
		ArtifactID:    "art-1",
		DaysCount:     7,
		AvgCalories7d: 2100,
	}

	result, err := repo.Upsert(context.Background(), summary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID == "" {
		t.Error("expected ID to be assigned")
	}
	if result.ClientID != "c1" {
		t.Errorf("expected client_id c1, got %s", result.ClientID)
	}

	all := repo.GetAll()
	if len(all) != 1 {
		t.Errorf("expected 1 summary, got %d", len(all))
	}
}

func TestInMemoryNutritionSummaryRepository_Upsert_UpdatesExisting(t *testing.T) {
	repo := NewInMemoryNutritionSummaryRepository()

	summary1 := &entities.NutritionSummary{
		ClientID:      "c1",
		ArtifactID:    "art-1",
		AvgCalories7d: 2000,
	}
	result1, _ := repo.Upsert(context.Background(), summary1)

	summary2 := &entities.NutritionSummary{
		ClientID:      "c1",
		ArtifactID:    "art-1",
		AvgCalories7d: 2200,
	}
	result2, err := repo.Upsert(context.Background(), summary2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should reuse the same ID
	if result2.ID != result1.ID {
		t.Errorf("expected same ID on update, got %s vs %s", result1.ID, result2.ID)
	}

	// Should have updated value
	if result2.AvgCalories7d != 2200 {
		t.Errorf("expected updated avg_calories_7d 2200, got %g", result2.AvgCalories7d)
	}

	// Should still have only 1 summary
	all := repo.GetAll()
	if len(all) != 1 {
		t.Errorf("expected 1 summary after upsert, got %d", len(all))
	}
}

func TestInMemoryNutritionSummaryRepository_Upsert_DifferentArtifacts_CreatesSeparate(t *testing.T) {
	repo := NewInMemoryNutritionSummaryRepository()

	_, _ = repo.Upsert(context.Background(), &entities.NutritionSummary{
		ClientID:   "c1",
		ArtifactID: "art-1",
	})
	_, _ = repo.Upsert(context.Background(), &entities.NutritionSummary{
		ClientID:   "c1",
		ArtifactID: "art-2",
	})

	all := repo.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 summaries, got %d", len(all))
	}
}
