package entities

import (
	"testing"
	"time"
)

func TestIsValidNutritionMetric_ValidMetrics_ReturnsTrue(t *testing.T) {
	metrics := []NutritionMetric{
		NutritionMetricCalories,
		NutritionMetricProtein,
		NutritionMetricCarbs,
		NutritionMetricFat,
		NutritionMetricFiber,
	}
	for _, m := range metrics {
		if !IsValidNutritionMetric(m) {
			t.Errorf("expected metric %q to be valid", m)
		}
	}
}

func TestIsValidNutritionMetric_InvalidMetric_ReturnsFalse(t *testing.T) {
	if IsValidNutritionMetric("invalid") {
		t.Error("expected invalid metric to return false")
	}
}

func TestNutritionUnit_AllMetricsHaveUnits(t *testing.T) {
	for _, m := range AllNutritionMetrics {
		unit, ok := NutritionUnit[m]
		if !ok {
			t.Errorf("metric %q has no unit mapping", m)
		}
		if unit == "" {
			t.Errorf("metric %q has empty unit", m)
		}
	}
}

func TestComputeDailyTotals_MultipleMeals_SumsCorrectly(t *testing.T) {
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	entries := []NutritionEntry{
		{Date: date, Meal: "Breakfast", Calories: 450, Protein: 35, Carbs: 48, Fat: 12, Fiber: 6},
		{Date: date, Meal: "Lunch", Calories: 680, Protein: 45, Carbs: 65, Fat: 22, Fiber: 8},
		{Date: date, Meal: "Dinner", Calories: 720, Protein: 40, Carbs: 70, Fat: 28, Fiber: 5},
		{Date: date, Meal: "Snack", Calories: 200, Protein: 12, Carbs: 20, Fat: 8, Fiber: 2},
	}

	totals := ComputeDailyTotals(entries)
	if len(totals) != 1 {
		t.Fatalf("expected 1 day, got %d", len(totals))
	}

	day := totals[0]
	assertFloat(t, "Calories", 2050, day.Calories)
	assertFloat(t, "Protein", 132, day.Protein)
	assertFloat(t, "Carbs", 203, day.Carbs)
	assertFloat(t, "Fat", 70, day.Fat)
	assertFloat(t, "Fiber", 21, day.Fiber)
}

func TestComputeDailyTotals_MultipleDays_GroupsCorrectly(t *testing.T) {
	day1 := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)

	entries := []NutritionEntry{
		{Date: day1, Meal: "Breakfast", Calories: 400, Protein: 30, Carbs: 50, Fat: 10, Fiber: 5},
		{Date: day2, Meal: "Breakfast", Calories: 500, Protein: 40, Carbs: 60, Fat: 15, Fiber: 7},
		{Date: day1, Meal: "Lunch", Calories: 600, Protein: 35, Carbs: 55, Fat: 20, Fiber: 6},
	}

	totals := ComputeDailyTotals(entries)
	if len(totals) != 2 {
		t.Fatalf("expected 2 days, got %d", len(totals))
	}

	// Day 1: 400+600=1000 cal, 30+35=65 protein, 50+55=105 carbs, 10+20=30 fat, 5+6=11 fiber
	assertFloat(t, "Day1 Calories", 1000, totals[0].Calories)
	assertFloat(t, "Day1 Protein", 65, totals[0].Protein)

	// Day 2: 500 cal, 40 protein
	assertFloat(t, "Day2 Calories", 500, totals[1].Calories)
	assertFloat(t, "Day2 Protein", 40, totals[1].Protein)
}

func TestComputeDailyTotals_EmptyInput_ReturnsNil(t *testing.T) {
	totals := ComputeDailyTotals(nil)
	if totals != nil {
		t.Errorf("expected nil, got %v", totals)
	}

	totals = ComputeDailyTotals([]NutritionEntry{})
	if totals != nil {
		t.Errorf("expected nil, got %v", totals)
	}
}

func TestComputeAverages_SevenDays_ComputesCorrectly(t *testing.T) {
	dailyData := make([]DailyNutrition, 7)
	for i := range dailyData {
		dailyData[i] = DailyNutrition{
			Calories: float64(2000 + i*100),
			Protein:  float64(150 + i*5),
			Carbs:    float64(200 + i*10),
			Fat:      float64(70 + i*2),
			Fiber:    float64(25 + i),
		}
	}

	avg := ComputeAverages(dailyData, 7)

	// Average of 2000,2100,2200,2300,2400,2500,2600 = 2300
	assertFloat(t, "Avg Calories", 2300, avg.Calories)
	// Average of 150,155,160,165,170,175,180 = 165
	assertFloat(t, "Avg Protein", 165, avg.Protein)
}

func TestComputeAverages_FewerDaysThanWindow_UsesAllAvailable(t *testing.T) {
	dailyData := []DailyNutrition{
		{Calories: 2000, Protein: 150, Carbs: 200, Fat: 70, Fiber: 25},
		{Calories: 2200, Protein: 160, Carbs: 220, Fat: 80, Fiber: 30},
	}

	avg := ComputeAverages(dailyData, 7)
	assertFloat(t, "Avg Calories", 2100, avg.Calories)
	assertFloat(t, "Avg Protein", 155, avg.Protein)
}

func TestComputeAverages_SingleDay_EqualsItself(t *testing.T) {
	dailyData := []DailyNutrition{
		{Calories: 2000, Protein: 150, Carbs: 200, Fat: 70, Fiber: 25},
	}

	avg := ComputeAverages(dailyData, 7)
	assertFloat(t, "Avg Calories", 2000, avg.Calories)
	assertFloat(t, "Avg Protein", 150, avg.Protein)
}

func TestComputeAverages_EmptyInput_ReturnsZero(t *testing.T) {
	avg := ComputeAverages(nil, 7)
	assertFloat(t, "Avg Calories", 0, avg.Calories)
}

func TestComputeAverages_ZeroWindow_ReturnsZero(t *testing.T) {
	dailyData := []DailyNutrition{
		{Calories: 2000},
	}
	avg := ComputeAverages(dailyData, 0)
	assertFloat(t, "Avg Calories", 0, avg.Calories)
}

func TestComputeAverages_TakesLastNDays(t *testing.T) {
	dailyData := []DailyNutrition{
		{Calories: 1000},
		{Calories: 2000},
		{Calories: 3000},
	}

	avg := ComputeAverages(dailyData, 2)
	// Should take last 2: 2000+3000 / 2 = 2500
	assertFloat(t, "Avg Calories", 2500, avg.Calories)
}

func assertFloat(t *testing.T, name string, expected, actual float64) {
	t.Helper()
	diff := expected - actual
	if diff < -0.01 || diff > 0.01 {
		t.Errorf("%s: expected %.2f, got %.2f", name, expected, actual)
	}
}
