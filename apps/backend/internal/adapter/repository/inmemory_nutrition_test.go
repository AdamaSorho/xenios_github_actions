package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryNutritionRepository_BatchCreateMeasurements_StoresMeasurements(t *testing.T) {
	repo := NewInMemoryNutritionRepository()
	measurements := []*entities.Measurement{
		{ClientID: "c1", RecordedBy: "co1", MeasurementType: "calories_kcal", Value: 2000, Unit: "kcal", MeasuredAt: time.Now()},
		{ClientID: "c1", RecordedBy: "co1", MeasurementType: "protein_g", Value: 150, Unit: "g", MeasuredAt: time.Now()},
	}

	err := repo.BatchCreateMeasurements(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored := repo.GetMeasurements()
	if len(stored) != 2 {
		t.Fatalf("expected 2 measurements, got %d", len(stored))
	}
	if stored[0].ID == "" {
		t.Error("expected ID to be generated")
	}
}

func TestInMemoryNutritionRepository_BatchCreateMeasurements_AssignsIDs(t *testing.T) {
	repo := NewInMemoryNutritionRepository()
	measurements := []*entities.Measurement{
		{ClientID: "c1", MeasurementType: "calories_kcal", Value: 2000, Unit: "kcal", MeasuredAt: time.Now()},
	}
	err := repo.BatchCreateMeasurements(context.Background(), measurements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	stored := repo.GetMeasurements()
	if stored[0].ID == "" {
		t.Error("expected auto-generated ID")
	}
}

func TestInMemoryNutritionRepository_StoreAverages_StoresAverages(t *testing.T) {
	repo := NewInMemoryNutritionRepository()
	averages := []*entities.NutritionAverage{
		{ClientID: "c1", PeriodDays: 7, AvgCalories: 2000, AvgProtein: 150},
		{ClientID: "c1", PeriodDays: 14, AvgCalories: 2100, AvgProtein: 155},
	}

	err := repo.StoreAverages(context.Background(), averages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored := repo.GetAverages()
	if len(stored) != 2 {
		t.Fatalf("expected 2 averages, got %d", len(stored))
	}
}

func TestInMemoryNutritionRepository_FindMeasurementsByClientAndType_FiltersCorrectly(t *testing.T) {
	repo := NewInMemoryNutritionRepository()
	measurements := []*entities.Measurement{
		{ClientID: "c1", MeasurementType: "calories_kcal", Value: 2000, Unit: "kcal", MeasuredAt: time.Now()},
		{ClientID: "c1", MeasurementType: "protein_g", Value: 150, Unit: "g", MeasuredAt: time.Now()},
		{ClientID: "c2", MeasurementType: "calories_kcal", Value: 1800, Unit: "kcal", MeasuredAt: time.Now()},
	}
	_ = repo.BatchCreateMeasurements(context.Background(), measurements)

	result, err := repo.FindMeasurementsByClientAndType(context.Background(), "c1", "calories_kcal", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 measurement, got %d", len(result))
	}
	if result[0].Value != 2000 {
		t.Errorf("expected value 2000, got %f", result[0].Value)
	}
}

func TestInMemoryNutritionRepository_FindMeasurementsByClientAndType_RespectsLimit(t *testing.T) {
	repo := NewInMemoryNutritionRepository()
	measurements := []*entities.Measurement{
		{ClientID: "c1", MeasurementType: "calories_kcal", Value: 2000, Unit: "kcal", MeasuredAt: time.Now()},
		{ClientID: "c1", MeasurementType: "calories_kcal", Value: 2100, Unit: "kcal", MeasuredAt: time.Now()},
		{ClientID: "c1", MeasurementType: "calories_kcal", Value: 2200, Unit: "kcal", MeasuredAt: time.Now()},
	}
	_ = repo.BatchCreateMeasurements(context.Background(), measurements)

	result, err := repo.FindMeasurementsByClientAndType(context.Background(), "c1", "calories_kcal", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 measurements (limited), got %d", len(result))
	}
}

func TestInMemoryNutritionRepository_FindAveragesByClient_FiltersCorrectly(t *testing.T) {
	repo := NewInMemoryNutritionRepository()
	averages := []*entities.NutritionAverage{
		{ClientID: "c1", PeriodDays: 7, AvgCalories: 2000},
		{ClientID: "c2", PeriodDays: 7, AvgCalories: 1800},
	}
	_ = repo.StoreAverages(context.Background(), averages)

	result, err := repo.FindAveragesByClient(context.Background(), "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 average, got %d", len(result))
	}
	if result[0].AvgCalories != 2000 {
		t.Errorf("expected avg 2000, got %f", result[0].AvgCalories)
	}
}
