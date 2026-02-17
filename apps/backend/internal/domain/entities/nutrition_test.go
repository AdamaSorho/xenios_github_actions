package entities

import (
	"math"
	"testing"
	"time"
)

func TestIsValidNutritionMetricType_ValidTypes_ReturnsTrue(t *testing.T) {
	validTypes := []NutritionMetricType{
		NutritionMetricCalories,
		NutritionMetricProtein,
		NutritionMetricCarbs,
		NutritionMetricFat,
		NutritionMetricFiber,
	}
	for _, mt := range validTypes {
		if !IsValidNutritionMetricType(mt) {
			t.Errorf("expected %q to be valid", mt)
		}
	}
}

func TestIsValidNutritionMetricType_InvalidType_ReturnsFalse(t *testing.T) {
	if IsValidNutritionMetricType("unknown") {
		t.Error("expected 'unknown' to be invalid")
	}
}

func TestNutritionMetricUnit_AllTypesHaveUnits(t *testing.T) {
	expectedUnits := map[NutritionMetricType]string{
		NutritionMetricCalories: "kcal",
		NutritionMetricProtein:  "g",
		NutritionMetricCarbs:    "g",
		NutritionMetricFat:      "g",
		NutritionMetricFiber:    "g",
	}
	for mt, expected := range expectedUnits {
		got, ok := NutritionMetricUnit[mt]
		if !ok {
			t.Errorf("missing unit for %q", mt)
			continue
		}
		if got != expected {
			t.Errorf("unit for %q: got %q, want %q", mt, got, expected)
		}
	}
}

func date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}

func TestComputeAverages_EmptyTotals_ReturnsZeroSummary(t *testing.T) {
	summary := ComputeAverages(nil)
	if summary.TotalDays != 0 {
		t.Errorf("expected 0 total days, got %d", summary.TotalDays)
	}
	if summary.AvgCalories7d != 0 {
		t.Errorf("expected 0 avg calories 7d, got %f", summary.AvgCalories7d)
	}
}

func TestComputeAverages_SingleDay_AllAveragesEqualDayTotals(t *testing.T) {
	totals := []NutritionDailyTotal{
		{Date: date(2026, 1, 15), Calories: 2000, Protein: 150, Carbs: 200, Fat: 80, Fiber: 25},
	}

	summary := ComputeAverages(totals)

	if summary.TotalDays != 1 {
		t.Errorf("expected 1 total day, got %d", summary.TotalDays)
	}
	if !almostEqual(summary.AvgCalories7d, 2000) {
		t.Errorf("avg calories 7d: got %f, want 2000", summary.AvgCalories7d)
	}
	if !almostEqual(summary.AvgProtein7d, 150) {
		t.Errorf("avg protein 7d: got %f, want 150", summary.AvgProtein7d)
	}
	if !almostEqual(summary.AvgCalories14d, 2000) {
		t.Errorf("avg calories 14d: got %f, want 2000", summary.AvgCalories14d)
	}
	if !almostEqual(summary.AvgCalories30d, 2000) {
		t.Errorf("avg calories 30d: got %f, want 2000", summary.AvgCalories30d)
	}
}

func TestComputeAverages_SevenDays_CorrectAverages(t *testing.T) {
	var totals []NutritionDailyTotal
	for i := 0; i < 7; i++ {
		totals = append(totals, NutritionDailyTotal{
			Date:     date(2026, 1, 15+i),
			Calories: float64(2000 + i*100), // 2000, 2100, ..., 2600
			Protein:  150,
			Carbs:    200,
			Fat:      80,
			Fiber:    25,
		})
	}

	summary := ComputeAverages(totals)

	if summary.TotalDays != 7 {
		t.Errorf("expected 7 total days, got %d", summary.TotalDays)
	}
	// Average of 2000..2600 = 2300
	if !almostEqual(summary.AvgCalories7d, 2300) {
		t.Errorf("avg calories 7d: got %f, want 2300", summary.AvgCalories7d)
	}
	// 7d, 14d, and 30d should all be the same when we only have 7 days
	if !almostEqual(summary.AvgCalories14d, 2300) {
		t.Errorf("avg calories 14d: got %f, want 2300", summary.AvgCalories14d)
	}
	if !almostEqual(summary.AvgCalories30d, 2300) {
		t.Errorf("avg calories 30d: got %f, want 2300", summary.AvgCalories30d)
	}
}

func TestComputeAverages_ThirtyDays_DifferentWindowsComputedCorrectly(t *testing.T) {
	var totals []NutritionDailyTotal
	for i := 0; i < 30; i++ {
		totals = append(totals, NutritionDailyTotal{
			Date:     date(2026, 1, 1+i),
			Calories: 2000,
			Protein:  150,
			Carbs:    200,
			Fat:      80,
			Fiber:    25,
		})
	}
	// Modify the last 7 days to have different values
	for i := 23; i < 30; i++ {
		totals[i].Calories = 2500
	}

	summary := ComputeAverages(totals)

	if summary.TotalDays != 30 {
		t.Errorf("expected 30 total days, got %d", summary.TotalDays)
	}
	// Last 7 days all have 2500 calories
	if !almostEqual(summary.AvgCalories7d, 2500) {
		t.Errorf("avg calories 7d: got %f, want 2500", summary.AvgCalories7d)
	}
	// Last 14 days: 7 days at 2000 + 7 days at 2500 = avg 2250
	if !almostEqual(summary.AvgCalories14d, 2250) {
		t.Errorf("avg calories 14d: got %f, want 2250", summary.AvgCalories14d)
	}
	// All 30 days: 23 at 2000 + 7 at 2500 = (46000 + 17500) / 30 = 2116.67
	expectedAvg30 := (23*2000.0 + 7*2500.0) / 30.0
	if !almostEqual(summary.AvgCalories30d, expectedAvg30) {
		t.Errorf("avg calories 30d: got %f, want %f", summary.AvgCalories30d, expectedAvg30)
	}
}
