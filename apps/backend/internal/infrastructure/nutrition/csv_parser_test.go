package nutrition

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
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

func TestCSVParser_Parse_EmptyData_ReturnsError(t *testing.T) {
	p := NewCSVParser()
	_, err := p.Parse([]byte{})
	if err == nil {
		t.Fatal("expected error for empty data")
	}
}

func TestCSVParser_Parse_InvalidHeaders_ReturnsError(t *testing.T) {
	p := NewCSVParser()
	_, err := p.Parse([]byte("name,age,city\nJohn,30,NYC\n"))
	if err == nil {
		t.Fatal("expected error for invalid headers")
	}
}

func TestCSVParser_Parse_GenericFormat_ParsesCorrectly(t *testing.T) {
	p := NewCSVParser()
	data := readTestFile(t, "generic.csv")
	logs, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("expected 3 daily logs, got %d", len(logs))
	}

	// Verify one entry
	found := false
	for _, log := range logs {
		if log.Date.Format("2006-01-02") == "2024-01-15" {
			found = true
			if log.Calories != 2150 {
				t.Errorf("expected calories 2150, got %f", log.Calories)
			}
			if log.Protein != 165 {
				t.Errorf("expected protein 165, got %f", log.Protein)
			}
			if log.Carbs != 220 {
				t.Errorf("expected carbs 220, got %f", log.Carbs)
			}
			if log.Fat != 85 {
				t.Errorf("expected fat 85, got %f", log.Fat)
			}
			if log.Fiber != 25 {
				t.Errorf("expected fiber 25, got %f", log.Fiber)
			}
		}
	}
	if !found {
		t.Error("expected to find entry for 2024-01-15")
	}
}

func TestCSVParser_Parse_MyFitnessPalFormat_ParsesCorrectly(t *testing.T) {
	p := NewCSVParser()
	data := readTestFile(t, "myfitnesspal.csv")
	logs, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("expected 3 daily logs, got %d", len(logs))
	}

	found := false
	for _, log := range logs {
		if log.Date.Format("2006-01-02") == "2024-01-15" {
			found = true
			if log.Calories != 2150 {
				t.Errorf("expected calories 2150, got %f", log.Calories)
			}
			if log.Protein != 165 {
				t.Errorf("expected protein 165, got %f", log.Protein)
			}
			if log.Carbs != 220 {
				t.Errorf("expected carbs 220, got %f", log.Carbs)
			}
			if log.Fat != 85 {
				t.Errorf("expected fat 85, got %f", log.Fat)
			}
			if log.Fiber != 25 {
				t.Errorf("expected fiber 25, got %f", log.Fiber)
			}
		}
	}
	if !found {
		t.Error("expected to find entry for 2024-01-15")
	}
}

func TestCSVParser_Parse_GenericNoFiber_ParsesWithZeroFiber(t *testing.T) {
	p := NewCSVParser()
	data := readTestFile(t, "generic_no_fiber.csv")
	logs, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 daily logs, got %d", len(logs))
	}
	for _, log := range logs {
		if log.Fiber != 0 {
			t.Errorf("expected fiber 0 for no-fiber format, got %f", log.Fiber)
		}
	}
}

func TestCSVParser_Parse_MultiMealPerDay_AggregatesByDate(t *testing.T) {
	p := NewCSVParser()
	data := readTestFile(t, "multi_meal.csv")
	logs, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 daily logs (aggregated), got %d", len(logs))
	}

	for _, log := range logs {
		switch log.Date.Format("2006-01-02") {
		case "2024-01-15":
			// 800 + 600 + 750 = 2150
			if log.Calories != 2150 {
				t.Errorf("expected aggregated calories 2150, got %f", log.Calories)
			}
			// 40 + 35 + 45 = 120
			if log.Protein != 120 {
				t.Errorf("expected aggregated protein 120, got %f", log.Protein)
			}
		case "2024-01-14":
			// 900 + 700 = 1600
			if log.Calories != 1600 {
				t.Errorf("expected aggregated calories 1600, got %f", log.Calories)
			}
		}
	}
}

func TestCSVParser_Parse_NoValidRows_ReturnsError(t *testing.T) {
	p := NewCSVParser()
	data := []byte("date,calories,protein,carbs,fat\nnot-a-date,abc,def,ghi,jkl\n")
	_, err := p.Parse(data)
	if err == nil {
		t.Fatal("expected error when no valid rows")
	}
}

func TestCSVParser_Parse_InlineData_ParsesCorrectly(t *testing.T) {
	p := NewCSVParser()
	data := []byte("date,calories,protein,carbs,fat,fiber\n2024-03-01,1800,140,180,70,20\n")
	logs, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
	if logs[0].Calories != 1800 {
		t.Errorf("expected 1800 calories, got %f", logs[0].Calories)
	}
}

func TestParseNumeric_CommaThousands_ParsesCorrectly(t *testing.T) {
	val, err := parseNumeric("2,150")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 2150 {
		t.Errorf("expected 2150, got %f", val)
	}
}

func TestParseNumeric_EmptyString_ReturnsZero(t *testing.T) {
	val, err := parseNumeric("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 0 {
		t.Errorf("expected 0, got %f", val)
	}
}

func TestParseNumeric_Dash_ReturnsZero(t *testing.T) {
	val, err := parseNumeric("--")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 0 {
		t.Errorf("expected 0, got %f", val)
	}
}

func TestParseDate_MultipleFormats_ParsesCorrectly(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2024-01-15", "2024-01-15"},
		{"01/15/2024", "2024-01-15"},
		{"1/5/2024", "2024-01-05"},
		{"2024/01/15", "2024-01-15"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			d, err := parseDate(tc.input)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.input, err)
			}
			if d.Format("2006-01-02") != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, d.Format("2006-01-02"))
			}
		})
	}
}

func TestParseDate_InvalidDate_ReturnsError(t *testing.T) {
	_, err := parseDate("not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestDetectFormat_GenericHeaders_ReturnsGeneric(t *testing.T) {
	headers := []string{"date", "calories", "protein", "carbs", "fat"}
	format, cm := detectFormat(headers)
	if format != FormatGeneric {
		t.Errorf("expected generic format, got %s", format)
	}
	if cm.date != 0 || cm.calories != 1 || cm.protein != 2 || cm.carbs != 3 || cm.fat != 4 {
		t.Errorf("unexpected column mapping: %+v", cm)
	}
}

func TestDetectFormat_MFPHeaders_ReturnsMFP(t *testing.T) {
	headers := []string{"date", "calories", "fat (g)", "carbohydrates (g)", "fiber (g)", "protein (g)"}
	format, cm := detectFormat(headers)
	if format != FormatMyFitnessPal {
		t.Errorf("expected myfitnesspal format, got %s", format)
	}
	if cm == nil {
		t.Fatal("expected column map")
	}
}

func TestDetectFormat_MissingDateColumn_ReturnsUnknown(t *testing.T) {
	headers := []string{"calories", "protein", "carbs"}
	format, _ := detectFormat(headers)
	if format != FormatUnknown {
		t.Errorf("expected unknown format, got %s", format)
	}
}

func TestDetectFormat_MissingCaloriesColumn_ReturnsUnknown(t *testing.T) {
	headers := []string{"date", "protein", "carbs"}
	format, _ := detectFormat(headers)
	if format != FormatUnknown {
		t.Errorf("expected unknown format, got %s", format)
	}
}
