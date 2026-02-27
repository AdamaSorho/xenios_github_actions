package parser

import (
	"os"
	"strings"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestCSVLabParser_Parse_StandardCSV_ExtractsAllMarkers(t *testing.T) {
	f, err := os.Open("testdata/lab_results_standard.csv")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	parser := NewCSVLabParser()
	measurements, err := parser.Parse(f)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(measurements) != 10 {
		t.Fatalf("got %d measurements, want 10", len(measurements))
	}

	// Verify specific markers
	byType := indexByType(measurements)

	assertMeasurement(t, byType, entities.LabTypeFastingGlucose, 98, "mg/dL", entities.LabFlagNormal)
	assertMeasurement(t, byType, entities.LabTypeTotalCholesterol, 210, "mg/dL", entities.LabFlagHigh)
	assertMeasurement(t, byType, entities.LabTypeLDLCholesterol, 142, "mg/dL", entities.LabFlagHigh)
	assertMeasurement(t, byType, entities.LabTypeHDLCholesterol, 55, "mg/dL", entities.LabFlagNormal)
	assertMeasurement(t, byType, entities.LabTypeTriglycerides, 120, "mg/dL", entities.LabFlagNormal)
	assertMeasurement(t, byType, entities.LabTypeHbA1c, 5.4, "%", entities.LabFlagNormal)
	assertMeasurement(t, byType, entities.LabTypeTestosterone, 450, "ng/dL", entities.LabFlagNormal)
	assertMeasurement(t, byType, entities.LabTypeTSH, 2.1, "mIU/L", entities.LabFlagNormal)
	assertMeasurement(t, byType, entities.LabTypeVitaminD, 32, "ng/mL", entities.LabFlagNormal)
	assertMeasurement(t, byType, entities.LabTypeIron, 85, "mcg/dL", entities.LabFlagNormal)
}

func TestCSVLabParser_Parse_OutOfRange_FlagsCorrectly(t *testing.T) {
	f, err := os.Open("testdata/lab_results_out_of_range.csv")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	parser := NewCSVLabParser()
	measurements, err := parser.Parse(f)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(measurements) != 5 {
		t.Fatalf("got %d measurements, want 5", len(measurements))
	}

	byType := indexByType(measurements)
	assertMeasurement(t, byType, entities.LabTypeFastingGlucose, 115, "mg/dL", entities.LabFlagHigh)
	assertMeasurement(t, byType, entities.LabTypeLDLCholesterol, 142, "mg/dL", entities.LabFlagHigh)
	assertMeasurement(t, byType, entities.LabTypeHDLCholesterol, 35, "mg/dL", entities.LabFlagLow)
	assertMeasurement(t, byType, entities.LabTypeTSH, 0.2, "mIU/L", entities.LabFlagLow)
	assertMeasurement(t, byType, entities.LabTypeVitaminD, 15, "ng/mL", entities.LabFlagLow)
}

func TestCSVLabParser_Parse_MissingRefRange_NilFlag(t *testing.T) {
	f, err := os.Open("testdata/lab_results_missing_ref.csv")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	parser := NewCSVLabParser()
	measurements, err := parser.Parse(f)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(measurements) != 2 {
		t.Fatalf("got %d measurements, want 2", len(measurements))
	}

	for _, m := range measurements {
		if m.Flag != nil {
			t.Errorf("marker %s: flag = %v, want nil", m.MeasurementType, *m.Flag)
		}
		if m.ReferenceLow != nil || m.ReferenceHigh != nil {
			t.Errorf("marker %s: reference range should be nil", m.MeasurementType)
		}
	}
}

func TestCSVLabParser_Parse_UnknownMarkers_Skipped(t *testing.T) {
	f, err := os.Open("testdata/lab_results_unknown_markers.csv")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	parser := NewCSVLabParser()
	measurements, err := parser.Parse(f)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Only the glucose marker should be extracted
	if len(measurements) != 1 {
		t.Fatalf("got %d measurements, want 1 (unknown markers skipped)", len(measurements))
	}
	if measurements[0].MeasurementType != entities.LabTypeFastingGlucose {
		t.Errorf("expected fasting_glucose, got %s", measurements[0].MeasurementType)
	}
}

func TestCSVLabParser_Parse_EmptyCSV_ReturnsEmpty(t *testing.T) {
	f, err := os.Open("testdata/lab_results_empty.csv")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	parser := NewCSVLabParser()
	measurements, err := parser.Parse(f)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(measurements) != 0 {
		t.Fatalf("got %d measurements, want 0", len(measurements))
	}
}

func TestCSVLabParser_Parse_BadFormat_ReturnsError(t *testing.T) {
	f, err := os.Open("testdata/lab_results_bad_format.csv")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	parser := NewCSVLabParser()
	_, err = parser.Parse(f)
	if err == nil {
		t.Fatal("expected error for bad format CSV")
	}
}

func TestCSVLabParser_Parse_DuplicateMarkers_BothStored(t *testing.T) {
	f, err := os.Open("testdata/lab_results_duplicate_markers.csv")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	parser := NewCSVLabParser()
	measurements, err := parser.Parse(f)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(measurements) != 2 {
		t.Fatalf("got %d measurements, want 2 (both glucose readings)", len(measurements))
	}

	if measurements[0].Value != 98 {
		t.Errorf("first glucose value = %v, want 98", measurements[0].Value)
	}
	if measurements[1].Value != 102 {
		t.Errorf("second glucose value = %v, want 102", measurements[1].Value)
	}
}

func TestParseReferenceRange_Range_ReturnsLowHigh(t *testing.T) {
	low, high := parseReferenceRange("70-100")
	if low == nil || *low != 70 {
		t.Errorf("low = %v, want 70", low)
	}
	if high == nil || *high != 100 {
		t.Errorf("high = %v, want 100", high)
	}
}

func TestParseReferenceRange_LessThan_ReturnsOnlyHigh(t *testing.T) {
	low, high := parseReferenceRange("<200")
	if low != nil {
		t.Errorf("low = %v, want nil", low)
	}
	if high == nil || *high != 200 {
		t.Errorf("high = %v, want 200", high)
	}
}

func TestParseReferenceRange_GreaterThan_ReturnsOnlyLow(t *testing.T) {
	low, high := parseReferenceRange(">40")
	if low == nil || *low != 40 {
		t.Errorf("low = %v, want 40", low)
	}
	if high != nil {
		t.Errorf("high = %v, want nil", high)
	}
}

func TestParseReferenceRange_Decimal_ParsesCorrectly(t *testing.T) {
	low, high := parseReferenceRange("0.4-4.0")
	if low == nil || *low != 0.4 {
		t.Errorf("low = %v, want 0.4", low)
	}
	if high == nil || *high != 4.0 {
		t.Errorf("high = %v, want 4.0", high)
	}
}

func TestParseReferenceRange_Empty_ReturnsNils(t *testing.T) {
	low, high := parseReferenceRange("")
	if low != nil || high != nil {
		t.Errorf("expected nils for empty range")
	}
}

func TestCSVLabParser_Parse_FromReader_Works(t *testing.T) {
	csvData := "Test Name,Result,Units,Reference Range\nGlucose,95,mg/dL,70-100\n"
	parser := NewCSVLabParser()
	measurements, err := parser.Parse(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(measurements) != 1 {
		t.Fatalf("got %d measurements, want 1", len(measurements))
	}
	if measurements[0].MeasurementType != entities.LabTypeFastingGlucose {
		t.Errorf("type = %s, want fasting_glucose", measurements[0].MeasurementType)
	}
}

func TestFindHeaderIndices_ValidHeaders_ReturnsIndices(t *testing.T) {
	header := []string{"Test Name", "Result", "Units", "Reference Range", "Flag"}
	idx, err := findHeaderIndices(header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx.testName != 0 || idx.result != 1 || idx.units != 2 || idx.refRange != 3 {
		t.Errorf("indices: testName=%d result=%d units=%d refRange=%d", idx.testName, idx.result, idx.units, idx.refRange)
	}
}

func TestFindHeaderIndices_MissingRequired_ReturnsError(t *testing.T) {
	header := []string{"Units", "Reference Range"}
	_, err := findHeaderIndices(header)
	if err == nil {
		t.Fatal("expected error for missing required columns")
	}
}

// indexByType creates a map from measurement type to the first measurement of that type.
func indexByType(measurements []entities.LabMeasurement) map[entities.LabMeasurementType]entities.LabMeasurement {
	m := make(map[entities.LabMeasurementType]entities.LabMeasurement)
	for _, measurement := range measurements {
		if _, exists := m[measurement.MeasurementType]; !exists {
			m[measurement.MeasurementType] = measurement
		}
	}
	return m
}

// assertMeasurement checks a measurement's value, unit, and flag.
func assertMeasurement(t *testing.T, byType map[entities.LabMeasurementType]entities.LabMeasurement, mt entities.LabMeasurementType, expectedValue float64, expectedUnit string, expectedFlag entities.LabFlag) {
	t.Helper()
	m, ok := byType[mt]
	if !ok {
		t.Errorf("missing measurement for %s", mt)
		return
	}
	if m.Value != expectedValue {
		t.Errorf("%s value = %v, want %v", mt, m.Value, expectedValue)
	}
	if m.Unit != expectedUnit {
		t.Errorf("%s unit = %q, want %q", mt, m.Unit, expectedUnit)
	}
	if m.Flag == nil {
		t.Errorf("%s flag is nil, want %s", mt, expectedFlag)
		return
	}
	if *m.Flag != expectedFlag {
		t.Errorf("%s flag = %s, want %s", mt, *m.Flag, expectedFlag)
	}
}
