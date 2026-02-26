package entities

import (
	"testing"
	"time"
)

func TestIsValidNutritionType_ValidTypes_ReturnsTrue(t *testing.T) {
	for _, nt := range AllNutritionTypes() {
		if !IsValidNutritionType(nt) {
			t.Errorf("expected %q to be a valid nutrition type", nt)
		}
	}
}

func TestIsValidNutritionType_InvalidType_ReturnsFalse(t *testing.T) {
	if IsValidNutritionType("invalid") {
		t.Error("expected 'invalid' to be an invalid nutrition type")
	}
}

func TestNutritionUnit_AllTypesHaveUnits(t *testing.T) {
	for _, nt := range AllNutritionTypes() {
		unit, ok := NutritionUnit[nt]
		if !ok {
			t.Errorf("nutrition type %q has no unit mapping", nt)
		}
		if unit == "" {
			t.Errorf("nutrition type %q has empty unit", nt)
		}
	}
}

func TestComputeDailyTotals_MultipleMeals_SumsCorrectly(t *testing.T) {
	day := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	rows := []NutritionRow{
		{Date: day, Meal: "Breakfast", Calories: 450, Protein: 35, Carbs: 48, Fat: 12, Fiber: 6},
		{Date: day, Meal: "Lunch", Calories: 680, Protein: 45, Carbs: 65, Fat: 22, Fiber: 8},
		{Date: day, Meal: "Dinner", Calories: 720, Protein: 40, Carbs: 70, Fat: 28, Fiber: 5},
		{Date: day, Meal: "Snack", Calories: 200, Protein: 12, Carbs: 20, Fat: 8, Fiber: 2},
	}

	totals := ComputeDailyTotals(rows)
	if len(totals) != 1 {
		t.Fatalf("expected 1 daily total, got %d", len(totals))
	}
	if totals[0].Calories != 2050 {
		t.Errorf("expected calories 2050, got %g", totals[0].Calories)
	}
	if totals[0].Protein != 132 {
		t.Errorf("expected protein 132, got %g", totals[0].Protein)
	}
	if totals[0].Carbs != 203 {
		t.Errorf("expected carbs 203, got %g", totals[0].Carbs)
	}
	if totals[0].Fat != 70 {
		t.Errorf("expected fat 70, got %g", totals[0].Fat)
	}
	if totals[0].Fiber != 21 {
		t.Errorf("expected fiber 21, got %g", totals[0].Fiber)
	}
}

func TestComputeDailyTotals_MultipleDays_GroupsByDate(t *testing.T) {
	day1 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)

	rows := []NutritionRow{
		{Date: day1, Meal: "Breakfast", Calories: 400, Protein: 30, Carbs: 50, Fat: 10, Fiber: 5},
		{Date: day2, Meal: "Breakfast", Calories: 500, Protein: 40, Carbs: 60, Fat: 15, Fiber: 7},
		{Date: day1, Meal: "Lunch", Calories: 600, Protein: 35, Carbs: 55, Fat: 20, Fiber: 6},
	}

	totals := ComputeDailyTotals(rows)
	if len(totals) != 2 {
		t.Fatalf("expected 2 daily totals, got %d", len(totals))
	}
	// Day 1: 400+600=1000
	if totals[0].Calories != 1000 {
		t.Errorf("day1: expected calories 1000, got %g", totals[0].Calories)
	}
	// Day 2: 500
	if totals[1].Calories != 500 {
		t.Errorf("day2: expected calories 500, got %g", totals[1].Calories)
	}
}

func TestComputeDailyTotals_EmptyInput_ReturnsEmpty(t *testing.T) {
	totals := ComputeDailyTotals(nil)
	if len(totals) != 0 {
		t.Errorf("expected 0 daily totals, got %d", len(totals))
	}
}

func TestComputeAverages_SevenDays_ComputesCorrectly(t *testing.T) {
	totals := make([]NutritionDailyTotal, 7)
	for i := range totals {
		totals[i] = NutritionDailyTotal{
			Date:     time.Date(2026, 1, 15+i, 0, 0, 0, 0, time.UTC),
			Calories: 2000,
			Protein:  150,
			Carbs:    200,
			Fat:      70,
			Fiber:    25,
		}
	}

	avg := ComputeAverages(totals, 7)
	if avg.Calories != 2000 {
		t.Errorf("expected avg calories 2000, got %g", avg.Calories)
	}
	if avg.Protein != 150 {
		t.Errorf("expected avg protein 150, got %g", avg.Protein)
	}
}

func TestComputeAverages_FewerDaysThanRequested_AveragesAllAvailable(t *testing.T) {
	totals := []NutritionDailyTotal{
		{Calories: 2000, Protein: 150, Carbs: 200, Fat: 70, Fiber: 25},
		{Calories: 2200, Protein: 160, Carbs: 220, Fat: 80, Fiber: 30},
	}

	avg := ComputeAverages(totals, 7) // asking for 7 but only 2 available
	expectedCal := (2000.0 + 2200.0) / 2
	if avg.Calories != expectedCal {
		t.Errorf("expected avg calories %g, got %g", expectedCal, avg.Calories)
	}
}

func TestComputeAverages_SingleDay_EqualsToThatDay(t *testing.T) {
	totals := []NutritionDailyTotal{
		{Calories: 2100, Protein: 165, Carbs: 220, Fat: 78, Fiber: 28},
	}

	avg := ComputeAverages(totals, 7)
	if avg.Calories != 2100 {
		t.Errorf("expected avg calories 2100, got %g", avg.Calories)
	}
	if avg.Protein != 165 {
		t.Errorf("expected avg protein 165, got %g", avg.Protein)
	}
}

func TestComputeAverages_EmptyInput_ReturnsZero(t *testing.T) {
	avg := ComputeAverages(nil, 7)
	if avg.Calories != 0 {
		t.Errorf("expected avg calories 0, got %g", avg.Calories)
	}
}

func TestComputeAverages_ThirtyDays_UsesLastThirty(t *testing.T) {
	totals := make([]NutritionDailyTotal, 60)
	for i := range totals {
		cal := float64(1000) // first 30 days
		if i >= 30 {
			cal = 3000 // last 30 days
		}
		totals[i] = NutritionDailyTotal{Calories: cal}
	}

	avg := ComputeAverages(totals, 30)
	if avg.Calories != 3000 {
		t.Errorf("expected avg calories 3000 (last 30 days), got %g", avg.Calories)
	}
}
