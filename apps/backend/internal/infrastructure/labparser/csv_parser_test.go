package labparser

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func testdataPath(filename string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "testdata", filename)
}

func TestParseCSV_QuestBasic_ExtractsAllMarkers(t *testing.T) {
	data, err := os.ReadFile(testdataPath("quest_basic.csv"))
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	results, err := ParseCSV(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 10 {
		t.Fatalf("expected 10 results, got %d", len(results))
	}

	// Verify first marker (glucose)
	glucose := findResult(results, string(entities.LabFastingGlucose))
	if glucose == nil {
		t.Fatal("glucose result not found")
	}
	if glucose.Value != 98 {
		t.Errorf("glucose value: expected 98, got %f", glucose.Value)
	}
	if glucose.Unit != "mg/dL" {
		t.Errorf("glucose unit: expected mg/dL, got %s", glucose.Unit)
	}
	if glucose.Flag != entities.FlagNormal {
		t.Errorf("glucose flag: expected normal, got %s", glucose.Flag)
	}
}

func TestParseCSV_QuestBasic_FlagsOutOfRange(t *testing.T) {
	data, err := os.ReadFile(testdataPath("quest_basic.csv"))
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	results, err := ParseCSV(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// LDL at 142 should be flagged high (reference < 100)
	ldl := findResult(results, string(entities.LabLDLCholesterol))
	if ldl == nil {
		t.Fatal("LDL result not found")
	}
	if ldl.Value != 142 {
		t.Errorf("LDL value: expected 142, got %f", ldl.Value)
	}
	if ldl.Flag != entities.FlagHigh {
		t.Errorf("LDL flag: expected high, got %s", ldl.Flag)
	}

	// Total cholesterol at 210 should be flagged high (reference < 200)
	totalChol := findResult(results, string(entities.LabTotalCholesterol))
	if totalChol == nil {
		t.Fatal("total cholesterol result not found")
	}
	if totalChol.Flag != entities.FlagHigh {
		t.Errorf("total cholesterol flag: expected high, got %s", totalChol.Flag)
	}
}

func TestParseCSV_LabcorpFormat_ExtractsMarkers(t *testing.T) {
	data, err := os.ReadFile(testdataPath("labcorp_basic.csv"))
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	results, err := ParseCSV(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 7 {
		t.Fatalf("expected 7 results, got %d", len(results))
	}

	// Fasting glucose at 105 should be flagged high (70-100)
	glucose := findResult(results, string(entities.LabFastingGlucose))
	if glucose == nil {
		t.Fatal("glucose result not found")
	}
	if glucose.Value != 105 {
		t.Errorf("glucose value: expected 105, got %f", glucose.Value)
	}
	if glucose.Flag != entities.FlagHigh {
		t.Errorf("glucose flag: expected high, got %s", glucose.Flag)
	}

	// TSH at 0.3 should be flagged low (0.4-4.0)
	tsh := findResult(results, string(entities.LabTSH))
	if tsh == nil {
		t.Fatal("TSH result not found")
	}
	if tsh.Flag != entities.FlagLow {
		t.Errorf("TSH flag: expected low, got %s", tsh.Flag)
	}

	// Vitamin D at 22 should be flagged low (30-100)
	vitD := findResult(results, string(entities.LabVitaminD))
	if vitD == nil {
		t.Fatal("vitamin D result not found")
	}
	if vitD.Flag != entities.FlagLow {
		t.Errorf("vitamin D flag: expected low, got %s", vitD.Flag)
	}
}

func TestParseCSV_MissingReference_FlagEmpty(t *testing.T) {
	data, err := os.ReadFile(testdataPath("missing_reference.csv"))
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	results, err := ParseCSV(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// LDL and Iron have no reference range
	ldl := findResult(results, string(entities.LabLDLCholesterol))
	if ldl == nil {
		t.Fatal("LDL result not found")
	}
	if ldl.Flag != "" {
		t.Errorf("LDL flag: expected empty (no reference), got %s", ldl.Flag)
	}
	if ldl.ReferenceLow != nil || ldl.ReferenceHigh != nil {
		t.Error("LDL reference range should be nil when missing")
	}
}

func TestParseCSV_EmptyCSV_ReturnsEmpty(t *testing.T) {
	data, err := os.ReadFile(testdataPath("empty.csv"))
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	results, err := ParseCSV(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results for empty CSV, got %d", len(results))
	}
}

func TestParseCSV_BadFormat_ReturnsError(t *testing.T) {
	data, err := os.ReadFile(testdataPath("bad_format.csv"))
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	_, err = ParseCSV(data)
	if err == nil {
		t.Error("expected error for bad format CSV, got nil")
	}
}

func TestParseCSV_DuplicateMarkers_ReturnsBoth(t *testing.T) {
	data, err := os.ReadFile(testdataPath("duplicate_markers.csv"))
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	results, err := ParseCSV(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results (duplicate glucose), got %d", len(results))
	}

	// Both should be glucose
	for _, r := range results {
		if r.MarkerName != string(entities.LabFastingGlucose) {
			t.Errorf("expected fasting_glucose, got %s", r.MarkerName)
		}
	}
}

func TestParseCSV_FromReader_Works(t *testing.T) {
	csv := "Test Name,Result,Units,Reference Range\nGlucose,95,mg/dL,70-100\n"
	results, err := ParseCSV([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestParseReferenceRange_Range(t *testing.T) {
	low, high := parseReferenceRange("70-100")
	if low == nil || *low != 70 {
		t.Errorf("expected low=70, got %v", low)
	}
	if high == nil || *high != 100 {
		t.Errorf("expected high=100, got %v", high)
	}
}

func TestParseReferenceRange_LessThan(t *testing.T) {
	low, high := parseReferenceRange("<200")
	if low != nil {
		t.Errorf("expected low=nil, got %v", low)
	}
	if high == nil || *high != 200 {
		t.Errorf("expected high=200, got %v", high)
	}
}

func TestParseReferenceRange_GreaterThan(t *testing.T) {
	low, high := parseReferenceRange(">40")
	if low == nil || *low != 40 {
		t.Errorf("expected low=40, got %v", low)
	}
	if high != nil {
		t.Errorf("expected high=nil, got %v", high)
	}
}

func TestParseReferenceRange_Decimal(t *testing.T) {
	low, high := parseReferenceRange("0.4-4.0")
	if low == nil || *low != 0.4 {
		t.Errorf("expected low=0.4, got %v", low)
	}
	if high == nil || *high != 4.0 {
		t.Errorf("expected high=4.0, got %v", high)
	}
}

func TestParseReferenceRange_Empty_ReturnsNil(t *testing.T) {
	low, high := parseReferenceRange("")
	if low != nil || high != nil {
		t.Errorf("expected nil/nil for empty, got %v/%v", low, high)
	}
}

func TestParseReferenceRange_LessThanDecimal(t *testing.T) {
	low, high := parseReferenceRange("<5.7")
	if low != nil {
		t.Errorf("expected low=nil, got %v", low)
	}
	if high == nil || *high != 5.7 {
		t.Errorf("expected high=5.7, got %v", high)
	}
}

func TestDetectCSVColumns_QuestFormat(t *testing.T) {
	header := []string{"Test Name", "Result", "Units", "Reference Range", "Flag"}
	cols, err := detectCSVColumns(header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cols.name < 0 || cols.value < 0 || cols.unit < 0 || cols.refRange < 0 {
		t.Errorf("failed to detect columns: %+v", cols)
	}
}

func TestDetectCSVColumns_LabcorpFormat(t *testing.T) {
	header := []string{"Test", "Value", "Unit", "Reference Range"}
	cols, err := detectCSVColumns(header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cols.name < 0 || cols.value < 0 || cols.unit < 0 || cols.refRange < 0 {
		t.Errorf("failed to detect columns: %+v", cols)
	}
}

func TestDetectCSVColumns_MissingRequiredColumns_ReturnsError(t *testing.T) {
	header := []string{"Date", "Time", "Location"}
	_, err := detectCSVColumns(header)
	if err == nil {
		t.Error("expected error for missing required columns")
	}
}

func TestNormalizeMarkerName_KnownMarkers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		found    bool
	}{
		{"Glucose, Fasting", string(entities.LabFastingGlucose), true},
		{"LDL Cholesterol", string(entities.LabLDLCholesterol), true},
		{"HDL-C", string(entities.LabHDLCholesterol), true},
		{"Unknown Marker", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name := strings.ToLower(strings.TrimSpace(tt.input))
			labType, ok := entities.KnownLabMarkers[name]
			if ok != tt.found {
				t.Errorf("expected found=%v, got %v", tt.found, ok)
			}
			if ok && string(labType) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, labType)
			}
		})
	}
}

func findResult(results []entities.LabResult, markerName string) *entities.LabResult {
	for i, r := range results {
		if r.MarkerName == markerName {
			return &results[i]
		}
	}
	return nil
}
