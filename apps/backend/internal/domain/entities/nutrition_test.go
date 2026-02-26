package entities

import (
	"testing"
	"time"
)

func TestIsValidNutritionType_ValidTypes_ReturnsTrue(t *testing.T) {
	validTypes := []NutritionMeasurementType{
		NutritionCalories,
		NutritionProtein,
		NutritionCarbs,
		NutritionFat,
		NutritionFiber,
	}
	for _, nt := range validTypes {
		if !IsValidNutritionType(nt) {
			t.Errorf("expected %s to be valid", nt)
		}
	}
}

func TestIsValidNutritionType_InvalidType_ReturnsFalse(t *testing.T) {
	if IsValidNutritionType("invalid_type") {
		t.Error("expected invalid_type to be invalid")
	}
}

func TestComputeNutritionAverages_EmptyLogs_ReturnsNil(t *testing.T) {
	result := ComputeNutritionAverages(nil, []int{7, 14, 30})
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestComputeNutritionAverages_SingleDay_ComputesCorrectly(t *testing.T) {
	logs := []NutritionDailyLog{
		{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Calories: 2000, Protein: 150, Carbs: 200, Fat: 80, Fiber: 30},
	}
	result := ComputeNutritionAverages(logs, []int{7})
	avg, ok := result[7]
	if !ok {
		t.Fatal("expected 7-day average")
	}
	if avg.AvgCalories != 2000 {
		t.Errorf("expected avg calories 2000, got %f", avg.AvgCalories)
	}
	if avg.AvgProtein != 150 {
		t.Errorf("expected avg protein 150, got %f", avg.AvgProtein)
	}
}

func TestComputeNutritionAverages_MultipleDays_ComputesCorrectly(t *testing.T) {
	logs := []NutritionDailyLog{
		{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Calories: 2000, Protein: 150, Carbs: 200, Fat: 80, Fiber: 30},
		{Date: time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC), Calories: 2200, Protein: 160, Carbs: 220, Fat: 90, Fiber: 25},
		{Date: time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC), Calories: 1800, Protein: 140, Carbs: 180, Fat: 70, Fiber: 35},
	}
	result := ComputeNutritionAverages(logs, []int{7, 14, 30})

	avg7, ok := result[7]
	if !ok {
		t.Fatal("expected 7-day average")
	}
	// With 3 days, averages over 3 days (capped at available data)
	expectedCal := (2000.0 + 2200.0 + 1800.0) / 3.0
	if avg7.AvgCalories != expectedCal {
		t.Errorf("expected avg calories %f, got %f", expectedCal, avg7.AvgCalories)
	}
	if avg7.PeriodDays != 7 {
		t.Errorf("expected period 7, got %d", avg7.PeriodDays)
	}
}

func TestComputeNutritionAverages_UsesRecentDaysFirst(t *testing.T) {
	// Create 10 days, only use 7 most recent
	logs := make([]NutritionDailyLog, 10)
	for i := 0; i < 10; i++ {
		logs[i] = NutritionDailyLog{
			Date:     time.Date(2024, 1, 10-i, 0, 0, 0, 0, time.UTC),
			Calories: float64(1000 + i*100),
			Protein:  100,
			Carbs:    200,
			Fat:      80,
			Fiber:    25,
		}
	}
	result := ComputeNutritionAverages(logs, []int{7})
	avg7 := result[7]

	// 7 most recent: days 10,9,8,7,6,5,4 → calories 1000,1100,1200,1300,1400,1500,1600
	expectedCal := (1000.0 + 1100 + 1200 + 1300 + 1400 + 1500 + 1600) / 7.0
	if avg7.AvgCalories != expectedCal {
		t.Errorf("expected avg calories %f, got %f", expectedCal, avg7.AvgCalories)
	}
}

func TestDailyLogsToMeasurements_SingleDay_Creates5Measurements(t *testing.T) {
	logs := []NutritionDailyLog{
		{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Calories: 2000, Protein: 150, Carbs: 200, Fat: 80, Fiber: 30},
	}
	measurements := DailyLogsToMeasurements(logs, "client-1", "coach-1", "artifact-1")
	if len(measurements) != 5 {
		t.Fatalf("expected 5 measurements, got %d", len(measurements))
	}
}

func TestDailyLogsToMeasurements_SetsCorrectFields(t *testing.T) {
	logs := []NutritionDailyLog{
		{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Calories: 2000, Protein: 150, Carbs: 200, Fat: 80, Fiber: 30},
	}
	measurements := DailyLogsToMeasurements(logs, "client-1", "coach-1", "artifact-1")

	// Check first measurement (calories)
	m := measurements[0]
	if m.ClientID != "client-1" {
		t.Errorf("expected client_id 'client-1', got '%s'", m.ClientID)
	}
	if m.RecordedBy != "coach-1" {
		t.Errorf("expected recorded_by 'coach-1', got '%s'", m.RecordedBy)
	}
	if m.MeasurementType != "calories_kcal" {
		t.Errorf("expected type 'calories_kcal', got '%s'", m.MeasurementType)
	}
	if m.Value != 2000 {
		t.Errorf("expected value 2000, got %f", m.Value)
	}
	if m.Unit != "kcal" {
		t.Errorf("expected unit 'kcal', got '%s'", m.Unit)
	}
	if m.SourceArtifactID != "artifact-1" {
		t.Errorf("expected artifact_id 'artifact-1', got '%s'", m.SourceArtifactID)
	}
}

func TestDailyLogsToMeasurements_MultipleDays_CorrectCount(t *testing.T) {
	logs := []NutritionDailyLog{
		{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Calories: 2000, Protein: 150, Carbs: 200, Fat: 80, Fiber: 30},
		{Date: time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC), Calories: 2200, Protein: 160, Carbs: 220, Fat: 90, Fiber: 25},
	}
	measurements := DailyLogsToMeasurements(logs, "client-1", "coach-1", "artifact-1")
	// 2 days * 5 types = 10
	if len(measurements) != 10 {
		t.Fatalf("expected 10 measurements, got %d", len(measurements))
	}
}

func TestDailyLogsToMeasurements_EmptyLogs_ReturnsEmpty(t *testing.T) {
	measurements := DailyLogsToMeasurements(nil, "c", "co", "a")
	if len(measurements) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(measurements))
	}
}
