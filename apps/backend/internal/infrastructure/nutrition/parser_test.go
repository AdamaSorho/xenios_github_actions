package nutrition

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func testdataPath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata", name)
}

func readTestFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(testdataPath(name))
	if err != nil {
		t.Fatalf("failed to read test file %s: %v", name, err)
	}
	return data
}

// --- DetectFormat tests ---

func TestDetectFormat_MyFitnessPalCSV_ReturnsMFP(t *testing.T) {
	data := readTestFile(t, "myfitnesspal_30days.csv")
	format := DetectFormat(data)
	if format != entities.CSVFormatMyFitnessPal {
		t.Errorf("expected %s, got %s", entities.CSVFormatMyFitnessPal, format)
	}
}

func TestDetectFormat_GenericCSV_ReturnsGeneric(t *testing.T) {
	data := readTestFile(t, "generic_nutrition.csv")
	format := DetectFormat(data)
	if format != entities.CSVFormatGeneric {
		t.Errorf("expected %s, got %s", entities.CSVFormatGeneric, format)
	}
}

func TestDetectFormat_EmptyInput_ReturnsUnknown(t *testing.T) {
	format := DetectFormat([]byte{})
	if format != entities.CSVFormatUnknown {
		t.Errorf("expected %s, got %s", entities.CSVFormatUnknown, format)
	}
}

func TestDetectFormat_InvalidCSV_ReturnsUnknown(t *testing.T) {
	format := DetectFormat([]byte("this is not a csv"))
	if format != entities.CSVFormatUnknown {
		t.Errorf("expected %s, got %s", entities.CSVFormatUnknown, format)
	}
}

// --- Parse tests ---

func TestParse_MyFitnessPal30Days_Returns30DailyTotals(t *testing.T) {
	data := readTestFile(t, "myfitnesspal_30days.csv")
	result, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Format != entities.CSVFormatMyFitnessPal {
		t.Errorf("expected format %s, got %s", entities.CSVFormatMyFitnessPal, result.Format)
	}
	if len(result.DailyTotals) != 30 {
		t.Errorf("expected 30 daily totals, got %d", len(result.DailyTotals))
	}
}

func TestParse_MyFitnessPalSingleDay_SumsMealsCorrectly(t *testing.T) {
	data := readTestFile(t, "single_day.csv")
	result, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.DailyTotals) != 1 {
		t.Fatalf("expected 1 daily total, got %d", len(result.DailyTotals))
	}

	day := result.DailyTotals[0]
	// 450+680+720+200 = 2050
	if day.Calories != 2050 {
		t.Errorf("expected calories 2050, got %.0f", day.Calories)
	}
	// 12+22+28+8 = 70
	if day.Fat != 70 {
		t.Errorf("expected fat 70, got %.0f", day.Fat)
	}
	// 35+45+40+12 = 132
	if day.Protein != 132 {
		t.Errorf("expected protein 132, got %.0f", day.Protein)
	}
	// 48+65+70+20 = 203
	if day.Carbs != 203 {
		t.Errorf("expected carbs 203, got %.0f", day.Carbs)
	}
	// 6+8+5+2 = 21
	if day.Fiber != 21 {
		t.Errorf("expected fiber 21, got %.0f", day.Fiber)
	}
}

func TestParse_GenericCSV_ParsesCorrectly(t *testing.T) {
	data := readTestFile(t, "generic_nutrition.csv")
	result, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Format != entities.CSVFormatGeneric {
		t.Errorf("expected format %s, got %s", entities.CSVFormatGeneric, result.Format)
	}
	if len(result.DailyTotals) != 3 {
		t.Errorf("expected 3 daily totals, got %d", len(result.DailyTotals))
	}
	// Verify first day values match exactly (no summing needed for generic)
	day := result.DailyTotals[0]
	if day.Calories != 2050 {
		t.Errorf("expected calories 2050, got %.0f", day.Calories)
	}
	if day.Protein != 132 {
		t.Errorf("expected protein 132, got %.0f", day.Protein)
	}
}

func TestParse_BadValues_SkipsInvalidRows(t *testing.T) {
	data := readTestFile(t, "bad_values.csv")
	result, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 2 rows skipped (abc calories, xyz fat), remaining: 3 valid for day 1 + 1 valid for day 2
	if result.SkippedRows != 2 {
		t.Errorf("expected 2 skipped rows, got %d", result.SkippedRows)
	}
	if len(result.DailyTotals) != 2 {
		t.Errorf("expected 2 days, got %d", len(result.DailyTotals))
	}
}

func TestParse_EmptyInput_ReturnsError(t *testing.T) {
	_, err := Parse([]byte{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParse_DailyTotalsSortedByDate(t *testing.T) {
	data := readTestFile(t, "myfitnesspal_30days.csv")
	result, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 1; i < len(result.DailyTotals); i++ {
		if result.DailyTotals[i].Date.Before(result.DailyTotals[i-1].Date) {
			t.Errorf("daily totals not sorted: %v before %v at index %d",
				result.DailyTotals[i].Date, result.DailyTotals[i-1].Date, i)
		}
	}
}

// --- ComputeAverages tests ---

func TestComputeAverages_30Days_AllPeriodsComputed(t *testing.T) {
	data := readTestFile(t, "myfitnesspal_30days.csv")
	result, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	avgs := ComputeAverages(result.DailyTotals)
	if avgs.Avg7Day == nil {
		t.Error("expected 7-day average to be computed")
	}
	if avgs.Avg14Day == nil {
		t.Error("expected 14-day average to be computed")
	}
	if avgs.Avg30Day == nil {
		t.Error("expected 30-day average to be computed")
	}
	if avgs.Avg7Day.Days != 7 {
		t.Errorf("expected 7-day average to cover 7 days, got %d", avgs.Avg7Day.Days)
	}
	if avgs.Avg30Day.Days != 30 {
		t.Errorf("expected 30-day average to cover 30 days, got %d", avgs.Avg30Day.Days)
	}
}

func TestComputeAverages_SingleDay_OnlyAllDayAverage(t *testing.T) {
	data := readTestFile(t, "single_day.csv")
	result, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	avgs := ComputeAverages(result.DailyTotals)
	// With 1 day, only the 7-day average is computed (it uses min(days, 7))
	if avgs.Avg7Day == nil {
		t.Error("expected 7-day average to be computed even with 1 day")
	}
	if avgs.Avg7Day.Days != 1 {
		t.Errorf("expected 7-day average to cover 1 day, got %d", avgs.Avg7Day.Days)
	}
	// 14 and 30 day should still be computed with available data
	if avgs.Avg14Day == nil {
		t.Error("expected 14-day average to be computed")
	}
	if avgs.Avg30Day == nil {
		t.Error("expected 30-day average to be computed")
	}
	// Single day: averages equal to that day's totals
	if avgs.Avg7Day.Calories != 2050 {
		t.Errorf("expected avg calories 2050, got %.0f", avgs.Avg7Day.Calories)
	}
}

func TestComputeAverages_3Days_AveragesMatchExpected(t *testing.T) {
	data := readTestFile(t, "generic_nutrition.csv")
	result, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	avgs := ComputeAverages(result.DailyTotals)
	if avgs.Avg7Day == nil {
		t.Fatal("expected 7-day average to be computed")
	}
	// (2050+1930+2020) / 3 = 2000
	expectedCalAvg := 2000.0
	if avgs.Avg7Day.Calories != expectedCalAvg {
		t.Errorf("expected avg calories %.0f, got %.0f", expectedCalAvg, avgs.Avg7Day.Calories)
	}
}

// --- DailyTotalsToMeasurements tests ---

func TestDailyTotalsToMeasurements_SingleDay_CreatesFiveMeasurements(t *testing.T) {
	daily := []entities.DailyNutrition{
		{
			Date:     time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			Calories: 2050,
			Protein:  132,
			Carbs:    203,
			Fat:      70,
			Fiber:    21,
		},
	}
	measurements := DailyTotalsToMeasurements(daily, "client-1", "coach-1")
	if len(measurements) != 5 {
		t.Errorf("expected 5 measurements, got %d", len(measurements))
	}

	// Verify each measurement type is present
	typeMap := make(map[entities.MeasurementType]float64)
	for _, m := range measurements {
		typeMap[m.MeasurementType] = m.Value
		if m.ClientID != "client-1" {
			t.Errorf("expected client_id 'client-1', got '%s'", m.ClientID)
		}
		if m.RecordedBy != "coach-1" {
			t.Errorf("expected recorded_by 'coach-1', got '%s'", m.RecordedBy)
		}
	}
	if typeMap[entities.MeasurementTypeCalories] != 2050 {
		t.Errorf("expected calories 2050, got %.0f", typeMap[entities.MeasurementTypeCalories])
	}
	if typeMap[entities.MeasurementTypeProtein] != 132 {
		t.Errorf("expected protein 132, got %.0f", typeMap[entities.MeasurementTypeProtein])
	}
	if typeMap[entities.MeasurementTypeCarbs] != 203 {
		t.Errorf("expected carbs 203, got %.0f", typeMap[entities.MeasurementTypeCarbs])
	}
	if typeMap[entities.MeasurementTypeFat] != 70 {
		t.Errorf("expected fat 70, got %.0f", typeMap[entities.MeasurementTypeFat])
	}
	if typeMap[entities.MeasurementTypeFiber] != 21 {
		t.Errorf("expected fiber 21, got %.0f", typeMap[entities.MeasurementTypeFiber])
	}
}

func TestDailyTotalsToMeasurements_MultipleDays_CreatesCorrectCount(t *testing.T) {
	daily := []entities.DailyNutrition{
		{Date: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), Calories: 2050, Protein: 132, Carbs: 203, Fat: 70, Fiber: 21},
		{Date: time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC), Calories: 1930, Protein: 120, Carbs: 196, Fat: 56, Fiber: 16},
	}
	measurements := DailyTotalsToMeasurements(daily, "client-1", "coach-1")
	// 2 days * 5 types = 10
	if len(measurements) != 10 {
		t.Errorf("expected 10 measurements, got %d", len(measurements))
	}
}

func TestDailyTotalsToMeasurements_CorrectUnits(t *testing.T) {
	daily := []entities.DailyNutrition{
		{Date: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), Calories: 2050, Protein: 132, Carbs: 203, Fat: 70, Fiber: 21},
	}
	measurements := DailyTotalsToMeasurements(daily, "client-1", "coach-1")
	for _, m := range measurements {
		expected := entities.MeasurementUnit(m.MeasurementType)
		if m.Unit != expected {
			t.Errorf("expected unit '%s' for type '%s', got '%s'", expected, m.MeasurementType, m.Unit)
		}
	}
}
