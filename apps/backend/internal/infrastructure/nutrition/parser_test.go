package nutrition

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func TestDetectFormat_MyFitnessPal_ReturnsFormatMFP(t *testing.T) {
	header := []string{"Date", "Meal", "Calories", "Fat (g)", "Protein (g)", "Carbs (g)", "Fiber (g)"}
	format := DetectFormat(header)
	if format != FormatMyFitnessPal {
		t.Errorf("expected FormatMyFitnessPal, got %v", format)
	}
}

func TestDetectFormat_Generic_ReturnsFormatGeneric(t *testing.T) {
	header := []string{"date", "calories", "protein", "carbs", "fat", "fiber"}
	format := DetectFormat(header)
	if format != FormatGeneric {
		t.Errorf("expected FormatGeneric, got %v", format)
	}
}

func TestDetectFormat_Unknown_ReturnsFormatUnknown(t *testing.T) {
	header := []string{"foo", "bar", "baz"}
	format := DetectFormat(header)
	if format != FormatUnknown {
		t.Errorf("expected FormatUnknown, got %v", format)
	}
}

func TestParseCSV_MyFitnessPalFixture_ParsesCorrectly(t *testing.T) {
	f, err := os.Open("testdata/myfitnesspal.csv")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	result, err := ParseCSV(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Format != FormatMyFitnessPal {
		t.Errorf("expected format MyFitnessPal, got %v", result.Format)
	}

	if len(result.DailyTotals) != 3 {
		t.Fatalf("expected 3 days, got %d", len(result.DailyTotals))
	}

	// Day 1: 450+680+720+200 = 2050 calories
	day1 := result.DailyTotals[0]
	if day1.Date != date(2026, 1, 15) {
		t.Errorf("day1 date: got %v, want 2026-01-15", day1.Date)
	}
	if day1.Calories != 2050 {
		t.Errorf("day1 calories: got %f, want 2050", day1.Calories)
	}
	// Protein: 35+45+40+12 = 132
	if day1.Protein != 132 {
		t.Errorf("day1 protein: got %f, want 132", day1.Protein)
	}
	// Fat: 12+22+28+8 = 70
	if day1.Fat != 70 {
		t.Errorf("day1 fat: got %f, want 70", day1.Fat)
	}
	// Carbs: 48+65+70+20 = 203
	if day1.Carbs != 203 {
		t.Errorf("day1 carbs: got %f, want 203", day1.Carbs)
	}
	// Fiber: 6+8+5+2 = 21
	if day1.Fiber != 21 {
		t.Errorf("day1 fiber: got %f, want 21", day1.Fiber)
	}

	if result.SkippedRows != 0 {
		t.Errorf("expected 0 skipped rows, got %d", result.SkippedRows)
	}
	if result.TotalRows != 11 {
		t.Errorf("expected 11 total rows, got %d", result.TotalRows)
	}
}

func TestParseCSV_GenericFixture_ParsesCorrectly(t *testing.T) {
	f, err := os.Open("testdata/generic.csv")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	result, err := ParseCSV(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Format != FormatGeneric {
		t.Errorf("expected format Generic, got %v", result.Format)
	}

	if len(result.DailyTotals) != 7 {
		t.Fatalf("expected 7 days, got %d", len(result.DailyTotals))
	}

	// First day should be 2026-01-15
	if result.DailyTotals[0].Date != date(2026, 1, 15) {
		t.Errorf("first day date: got %v, want 2026-01-15", result.DailyTotals[0].Date)
	}
	if result.DailyTotals[0].Calories != 2050 {
		t.Errorf("first day calories: got %f, want 2050", result.DailyTotals[0].Calories)
	}
}

func TestParseCSV_BadValues_SkipsInvalidRows(t *testing.T) {
	f, err := os.Open("testdata/bad_values.csv")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	result, err := ParseCSV(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 2 of 5 rows have bad numeric values, should be skipped
	if result.SkippedRows != 2 {
		t.Errorf("expected 2 skipped rows, got %d", result.SkippedRows)
	}

	// Should still have 2 days of valid data
	if len(result.DailyTotals) != 2 {
		t.Fatalf("expected 2 days, got %d", len(result.DailyTotals))
	}

	// Day 1 (2026-01-15): rows 1+3 (skip row 2 with "abc")
	day1 := result.DailyTotals[0]
	if day1.Calories != 1170 { // 450 + 720
		t.Errorf("day1 calories: got %f, want 1170", day1.Calories)
	}
}

func TestParseCSV_SingleDay_ParsesCorrectly(t *testing.T) {
	f, err := os.Open("testdata/single_day.csv")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	result, err := ParseCSV(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.DailyTotals) != 1 {
		t.Fatalf("expected 1 day, got %d", len(result.DailyTotals))
	}

	day := result.DailyTotals[0]
	if day.Calories != 2050 {
		t.Errorf("calories: got %f, want 2050", day.Calories)
	}
	if day.Protein != 132 {
		t.Errorf("protein: got %f, want 132", day.Protein)
	}
}

func TestParseCSV_EmptyInput_ReturnsError(t *testing.T) {
	r := strings.NewReader("")
	_, err := ParseCSV(r)
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseCSV_HeaderOnly_ReturnsEmptyTotals(t *testing.T) {
	r := strings.NewReader("date,calories,protein,carbs,fat,fiber\n")
	result, err := ParseCSV(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.DailyTotals) != 0 {
		t.Errorf("expected 0 daily totals, got %d", len(result.DailyTotals))
	}
}

func TestParseCSV_MultipleMealsPerDay_SumsCorrectly(t *testing.T) {
	csv := `Date,Meal,Calories,Fat (g),Protein (g),Carbs (g),Fiber (g)
2026-01-15,Breakfast,500,10,30,50,5
2026-01-15,Lunch,700,20,40,60,8
2026-01-15,Dinner,800,25,45,70,7
2026-01-15,Snack,200,5,10,20,3
`
	result, err := ParseCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.DailyTotals) != 1 {
		t.Fatalf("expected 1 day, got %d", len(result.DailyTotals))
	}

	day := result.DailyTotals[0]
	if day.Calories != 2200 {
		t.Errorf("calories: got %f, want 2200", day.Calories)
	}
	if day.Protein != 125 {
		t.Errorf("protein: got %f, want 125", day.Protein)
	}
	if day.Carbs != 200 {
		t.Errorf("carbs: got %f, want 200", day.Carbs)
	}
	if day.Fat != 60 {
		t.Errorf("fat: got %f, want 60", day.Fat)
	}
	if day.Fiber != 23 {
		t.Errorf("fiber: got %f, want 23", day.Fiber)
	}
}

func TestParseCSV_DailyTotalsSortedByDate(t *testing.T) {
	csv := `date,calories,protein,carbs,fat,fiber
2026-01-17,2000,150,200,80,25
2026-01-15,1900,140,190,70,20
2026-01-16,2100,160,210,85,28
`
	result, err := ParseCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.DailyTotals) != 3 {
		t.Fatalf("expected 3 days, got %d", len(result.DailyTotals))
	}

	if result.DailyTotals[0].Date != date(2026, 1, 15) {
		t.Errorf("first day should be 2026-01-15, got %v", result.DailyTotals[0].Date)
	}
	if result.DailyTotals[1].Date != date(2026, 1, 16) {
		t.Errorf("second day should be 2026-01-16, got %v", result.DailyTotals[1].Date)
	}
	if result.DailyTotals[2].Date != date(2026, 1, 17) {
		t.Errorf("third day should be 2026-01-17, got %v", result.DailyTotals[2].Date)
	}
}

func TestToNutritionRecords_ConvertsCorrectly(t *testing.T) {
	dailyTotals := []entities.NutritionDailyTotal{
		{Date: date(2026, 1, 15), Calories: 2000, Protein: 150, Carbs: 200, Fat: 80, Fiber: 25},
	}

	records := ToNutritionRecords(dailyTotals, "client-1", "coach-1", "artifact-1")

	// Should create 5 records (one per metric type)
	if len(records) != 5 {
		t.Fatalf("expected 5 records, got %d", len(records))
	}

	// Check that all metric types are present
	types := make(map[entities.NutritionMetricType]bool)
	for _, r := range records {
		types[r.MetricType] = true
		if r.ClientID != "client-1" {
			t.Errorf("expected client_id 'client-1', got '%s'", r.ClientID)
		}
		if r.CoachID != "coach-1" {
			t.Errorf("expected coach_id 'coach-1', got '%s'", r.CoachID)
		}
		if r.ArtifactID != "artifact-1" {
			t.Errorf("expected artifact_id 'artifact-1', got '%s'", r.ArtifactID)
		}
	}

	expectedTypes := []entities.NutritionMetricType{
		entities.NutritionMetricCalories,
		entities.NutritionMetricProtein,
		entities.NutritionMetricCarbs,
		entities.NutritionMetricFat,
		entities.NutritionMetricFiber,
	}
	for _, mt := range expectedTypes {
		if !types[mt] {
			t.Errorf("missing metric type: %s", mt)
		}
	}
}
