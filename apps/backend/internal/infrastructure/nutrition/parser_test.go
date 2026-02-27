package nutrition

import (
	"os"
	"testing"
)

func readTestFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("failed to read test file %s: %v", name, err)
	}
	return data
}

func TestDetectFormat_MFPHeaders_ReturnsMFP(t *testing.T) {
	data := readTestFile(t, "mfp_sample.csv")
	format := DetectFormat(data)
	if format != FormatMFP {
		t.Errorf("expected format %q, got %q", FormatMFP, format)
	}
}

func TestDetectFormat_GenericHeaders_ReturnsGeneric(t *testing.T) {
	data := readTestFile(t, "generic_sample.csv")
	format := DetectFormat(data)
	if format != FormatGeneric {
		t.Errorf("expected format %q, got %q", FormatGeneric, format)
	}
}

func TestDetectFormat_EmptyInput_ReturnsUnknown(t *testing.T) {
	format := DetectFormat(nil)
	if format != FormatUnknown {
		t.Errorf("expected format %q, got %q", FormatUnknown, format)
	}
}

func TestDetectFormat_GarbageInput_ReturnsUnknown(t *testing.T) {
	format := DetectFormat([]byte("this is not a csv"))
	if format != FormatUnknown {
		t.Errorf("expected format %q, got %q", FormatUnknown, format)
	}
}

func TestParse_MFPFormat_ParsesAllEntries(t *testing.T) {
	data := readTestFile(t, "mfp_sample.csv")
	entries, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 7 data rows (4 meals on day1 + 3 meals on day2)
	if len(entries) != 7 {
		t.Fatalf("expected 7 entries, got %d", len(entries))
	}

	// Check first entry
	e := entries[0]
	if e.Meal != "Breakfast" {
		t.Errorf("expected meal 'Breakfast', got %q", e.Meal)
	}
	assertFloat64(t, "Calories", 450, e.Calories)
	assertFloat64(t, "Protein", 35, e.Protein)
	assertFloat64(t, "Carbs", 48, e.Carbs)
	assertFloat64(t, "Fat", 12, e.Fat)
	assertFloat64(t, "Fiber", 6, e.Fiber)
}

func TestParse_MFPFormat_CorrectDate(t *testing.T) {
	data := readTestFile(t, "mfp_sample.csv")
	entries, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dateStr := entries[0].Date.Format("2006-01-02")
	if dateStr != "2026-01-15" {
		t.Errorf("expected date '2026-01-15', got %q", dateStr)
	}
}

func TestParse_GenericFormat_ParsesAllEntries(t *testing.T) {
	data := readTestFile(t, "generic_sample.csv")
	entries, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	e := entries[0]
	assertFloat64(t, "Calories", 2050, e.Calories)
	assertFloat64(t, "Protein", 132, e.Protein)
	assertFloat64(t, "Carbs", 203, e.Carbs)
	assertFloat64(t, "Fat", 70, e.Fat)
	assertFloat64(t, "Fiber", 21, e.Fiber)
}

func TestParse_BadValues_SkipsInvalidRows(t *testing.T) {
	data := readTestFile(t, "bad_values.csv")
	entries, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should skip the row with "bad_value" but parse the other 2
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries (1 skipped), got %d", len(entries))
	}
}

func TestParse_EmptyInput_ReturnsError(t *testing.T) {
	_, err := Parse(nil)
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParse_HeaderOnly_ReturnsEmptySlice(t *testing.T) {
	data := []byte("Date,Meal,Calories,Fat (g),Protein (g),Carbs (g),Fiber (g)\n")
	entries, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestParse_SingleDay_ReturnsSingleEntry(t *testing.T) {
	data := readTestFile(t, "single_day.csv")
	entries, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	assertFloat64(t, "Calories", 2050, entries[0].Calories)
}

func assertFloat64(t *testing.T, name string, expected, actual float64) {
	t.Helper()
	diff := expected - actual
	if diff < -0.01 || diff > 0.01 {
		t.Errorf("%s: expected %.2f, got %.2f", name, expected, actual)
	}
}
