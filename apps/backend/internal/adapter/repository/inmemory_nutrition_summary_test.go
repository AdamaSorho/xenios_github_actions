package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryNutritionSummaryRepository_Upsert_CreatesSummary(t *testing.T) {
	repo := NewInMemoryNutritionSummaryRepository()

	summary := &entities.NutritionSummary{
		ClientID:      "c1",
		ArtifactID:    "a1",
		AvgCalories7d: 2000,
		AvgProtein7d:  150,
		TotalDays:     7,
	}

	err := repo.Upsert(context.Background(), summary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results := repo.GetByClientID("c1")
	if len(results) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(results))
	}
	if results[0].AvgCalories7d != 2000 {
		t.Errorf("expected avg calories 2000, got %.2f", results[0].AvgCalories7d)
	}
}

func TestInMemoryNutritionSummaryRepository_Upsert_UpdatesExisting(t *testing.T) {
	repo := NewInMemoryNutritionSummaryRepository()

	summary := &entities.NutritionSummary{
		ClientID:      "c1",
		ArtifactID:    "a1",
		AvgCalories7d: 2000,
	}
	err := repo.Upsert(context.Background(), summary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated := &entities.NutritionSummary{
		ClientID:      "c1",
		ArtifactID:    "a1",
		AvgCalories7d: 2200,
	}
	err = repo.Upsert(context.Background(), updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	results := repo.GetByClientID("c1")
	if len(results) != 1 {
		t.Fatalf("expected 1 summary after upsert, got %d", len(results))
	}
	if results[0].AvgCalories7d != 2200 {
		t.Errorf("expected updated avg calories 2200, got %.2f", results[0].AvgCalories7d)
	}
}

func TestInMemoryNutritionSummaryRepository_GetByClientID_NoResults_ReturnsNil(t *testing.T) {
	repo := NewInMemoryNutritionSummaryRepository()

	results := repo.GetByClientID("nonexistent")
	if results != nil {
		t.Errorf("expected nil, got %v", results)
	}
}
