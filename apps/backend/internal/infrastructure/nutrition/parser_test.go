package nutrition

import (
	"testing"
	"time"
)

func TestParseCSV_MFPFormat_ParsesAllRows(t *testing.T) {
	csv := `Date,Meal,Calories,Fat (g),Protein (g),Carbs (g),Fiber (g)
2026-01-15,Breakfast,450,12,35,48,6
2026-01-15,Lunch,680,22,45,65,8
2026-01-15,Dinner,720,28,40,70,5
2026-01-15,Snack,200,8,12,20,2
`
	result, err := ParseCSV([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RowsParsed != 4 {
		t.Errorf("expected 4 rows parsed, got %d", result.RowsParsed)
	}
	if len(result.Rows) != 4 {
		t.Errorf("expected 4 rows, got %d", len(result.Rows))
	}
	if result.RowsSkipped != 0 {
		t.Errorf("expected 0 rows skipped, got %d", result.RowsSkipped)
	}
}

func TestParseCSV_MFPFormat_ParsesValues(t *testing.T) {
	csv := `Date,Meal,Calories,Fat (g),Protein (g),Carbs (g),Fiber (g)
2026-01-15,Breakfast,450,12,35,48,6
`
	result, err := ParseCSV([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}
	row := result.Rows[0]
	expectedDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	if !row.Date.Equal(expectedDate) {
		t.Errorf("expected date %v, got %v", expectedDate, row.Date)
	}
	if row.Meal != "Breakfast" {
		t.Errorf("expected meal Breakfast, got %s", row.Meal)
	}
	if row.Calories != 450 {
		t.Errorf("expected calories 450, got %g", row.Calories)
	}
	if row.Fat != 12 {
		t.Errorf("expected fat 12, got %g", row.Fat)
	}
	if row.Protein != 35 {
		t.Errorf("expected protein 35, got %g", row.Protein)
	}
	if row.Carbs != 48 {
		t.Errorf("expected carbs 48, got %g", row.Carbs)
	}
	if row.Fiber != 6 {
		t.Errorf("expected fiber 6, got %g", row.Fiber)
	}
}

func TestParseCSV_GenericFormat_ParsesCorrectly(t *testing.T) {
	csv := `date,calories,protein,carbs,fat,fiber
2026-01-15,2050,132,203,70,21
2026-01-16,2200,150,210,75,25
`
	result, err := ParseCSV([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RowsParsed != 2 {
		t.Errorf("expected 2 rows parsed, got %d", result.RowsParsed)
	}
	if len(result.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(result.Rows))
	}
	if result.Format != FormatGeneric {
		t.Errorf("expected format %q, got %q", FormatGeneric, result.Format)
	}
}

func TestParseCSV_GenericFormatWithoutFiber_ParsesCorrectly(t *testing.T) {
	csv := `date,calories,protein,carbs,fat
2026-01-15,2050,132,203,70
`
	result, err := ParseCSV([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RowsParsed != 1 {
		t.Errorf("expected 1 row parsed, got %d", result.RowsParsed)
	}
	if result.Rows[0].Fiber != 0 {
		t.Errorf("expected fiber 0, got %g", result.Rows[0].Fiber)
	}
}

func TestParseCSV_NonNumericValues_SkipsRow(t *testing.T) {
	csv := `Date,Meal,Calories,Fat (g),Protein (g),Carbs (g),Fiber (g)
2026-01-15,Breakfast,450,12,35,48,6
2026-01-15,Lunch,abc,22,45,65,8
2026-01-15,Dinner,720,28,40,70,5
`
	result, err := ParseCSV([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RowsParsed != 2 {
		t.Errorf("expected 2 rows parsed, got %d", result.RowsParsed)
	}
	if result.RowsSkipped != 1 {
		t.Errorf("expected 1 row skipped, got %d", result.RowsSkipped)
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
}

func TestParseCSV_InvalidDate_SkipsRow(t *testing.T) {
	csv := `date,calories,protein,carbs,fat
not-a-date,2050,132,203,70
2026-01-16,2200,150,210,75
`
	result, err := ParseCSV([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RowsParsed != 1 {
		t.Errorf("expected 1 row parsed, got %d", result.RowsParsed)
	}
	if result.RowsSkipped != 1 {
		t.Errorf("expected 1 row skipped, got %d", result.RowsSkipped)
	}
}

func TestParseCSV_EmptyInput_ReturnsError(t *testing.T) {
	_, err := ParseCSV([]byte(""))
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseCSV_HeaderOnly_ReturnsEmptyRows(t *testing.T) {
	csv := `Date,Meal,Calories,Fat (g),Protein (g),Carbs (g),Fiber (g)
`
	result, err := ParseCSV([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RowsParsed != 0 {
		t.Errorf("expected 0 rows parsed, got %d", result.RowsParsed)
	}
}

func TestParseCSV_UnrecognizedFormat_ReturnsError(t *testing.T) {
	csv := `name,age,city
John,30,NYC
`
	_, err := ParseCSV([]byte(csv))
	if err == nil {
		t.Fatal("expected error for unrecognized format")
	}
}

func TestParseCSV_MFPFormat_DetectsFormat(t *testing.T) {
	csv := `Date,Meal,Calories,Fat (g),Protein (g),Carbs (g),Fiber (g)
2026-01-15,Breakfast,450,12,35,48,6
`
	result, err := ParseCSV([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Format != FormatMyFitnessPal {
		t.Errorf("expected format %q, got %q", FormatMyFitnessPal, result.Format)
	}
}

func TestParseCSV_MultipleDays_ParsesAll(t *testing.T) {
	csv := `Date,Meal,Calories,Fat (g),Protein (g),Carbs (g),Fiber (g)
2026-01-15,Breakfast,450,12,35,48,6
2026-01-16,Breakfast,500,15,40,50,7
2026-01-17,Breakfast,480,13,38,45,5
`
	result, err := ParseCSV([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RowsParsed != 3 {
		t.Errorf("expected 3 rows parsed, got %d", result.RowsParsed)
	}

	// Verify different dates
	dates := make(map[string]bool)
	for _, r := range result.Rows {
		dates[r.Date.Format("2006-01-02")] = true
	}
	if len(dates) != 3 {
		t.Errorf("expected 3 unique dates, got %d", len(dates))
	}
}

func TestDetectFormat_MFPHeaders_ReturnsMyFitnessPal(t *testing.T) {
	headers := []string{"Date", "Meal", "Calories", "Fat (g)", "Protein (g)", "Carbs (g)", "Fiber (g)"}
	format := detectFormat(headers)
	if format != FormatMyFitnessPal {
		t.Errorf("expected %q, got %q", FormatMyFitnessPal, format)
	}
}

func TestDetectFormat_GenericHeaders_ReturnsGeneric(t *testing.T) {
	headers := []string{"date", "calories", "protein", "carbs", "fat"}
	format := detectFormat(headers)
	if format != FormatGeneric {
		t.Errorf("expected %q, got %q", FormatGeneric, format)
	}
}

func TestDetectFormat_UnknownHeaders_ReturnsUnknown(t *testing.T) {
	headers := []string{"name", "age", "city"}
	format := detectFormat(headers)
	if format != FormatUnknown {
		t.Errorf("expected %q, got %q", FormatUnknown, format)
	}
}
