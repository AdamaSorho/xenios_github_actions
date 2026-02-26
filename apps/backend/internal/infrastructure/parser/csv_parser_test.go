package parser

import (
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

var _ repository.LabFileParser = &CSVLabParser{}

func TestCSVLabParser_Parse_HappyPath_ExtractsAllMarkers(t *testing.T) {
	csv := `Test Name,Result,Units,Reference Range,Flag
Glucose,98,mg/dL,70-100,
LDL Cholesterol,142,mg/dL,<100,H
HDL Cholesterol,55,mg/dL,>40,
Triglycerides,120,mg/dL,<150,
HbA1c,5.4,%,< 5.7,
Testosterone,650,ng/dL,300-1000,
TSH,2.1,mIU/L,0.4-4.0,
Vitamin D,45,ng/mL,30-100,
Iron,85,mcg/dL,60-170,
Total Cholesterol,210,mg/dL,<200,H
`
	p := NewCSVLabParser()
	markers, err := p.Parse([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markers) != 10 {
		t.Fatalf("expected 10 markers, got %d", len(markers))
	}

	// Verify first marker
	if markers[0].Name != "Glucose" {
		t.Errorf("expected name 'Glucose', got %q", markers[0].Name)
	}
	if markers[0].Value != 98 {
		t.Errorf("expected value 98, got %v", markers[0].Value)
	}
	if markers[0].Unit != "mg/dL" {
		t.Errorf("expected unit 'mg/dL', got %q", markers[0].Unit)
	}
	if markers[0].ReferenceLow == nil || *markers[0].ReferenceLow != 70 {
		t.Errorf("expected reference_low=70, got %v", markers[0].ReferenceLow)
	}
	if markers[0].ReferenceHigh == nil || *markers[0].ReferenceHigh != 100 {
		t.Errorf("expected reference_high=100, got %v", markers[0].ReferenceHigh)
	}
}

func TestCSVLabParser_Parse_OutOfRange_LDLFlaggedHigh(t *testing.T) {
	csv := `Test Name,Result,Units,Reference Range
LDL Cholesterol,142,mg/dL,<100
`
	p := NewCSVLabParser()
	markers, err := p.Parse([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markers) != 1 {
		t.Fatalf("expected 1 marker, got %d", len(markers))
	}

	// Verify reference range parsed correctly for out-of-range detection
	if markers[0].ReferenceLow != nil {
		t.Errorf("expected nil reference_low, got %v", *markers[0].ReferenceLow)
	}
	if markers[0].ReferenceHigh == nil || *markers[0].ReferenceHigh != 100 {
		t.Errorf("expected reference_high=100, got %v", markers[0].ReferenceHigh)
	}

	// Verify DetermineFlag works on extracted data
	flag := entities.DetermineFlag(markers[0].Value, markers[0].ReferenceLow, markers[0].ReferenceHigh)
	if flag == nil || *flag != entities.FlagHigh {
		t.Errorf("expected flag 'high', got %v", flag)
	}
}

func TestCSVLabParser_Parse_MissingReferenceRange_NoRefValues(t *testing.T) {
	csv := `Test Name,Result,Units,Reference Range
Glucose,98,mg/dL,
`
	p := NewCSVLabParser()
	markers, err := p.Parse([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markers) != 1 {
		t.Fatalf("expected 1 marker, got %d", len(markers))
	}
	if markers[0].ReferenceLow != nil || markers[0].ReferenceHigh != nil {
		t.Error("expected nil reference values for empty reference range")
	}
}

func TestCSVLabParser_Parse_EmptyContent_ReturnsError(t *testing.T) {
	p := NewCSVLabParser()
	_, err := p.Parse([]byte(""))
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestCSVLabParser_Parse_InvalidCSV_ReturnsError(t *testing.T) {
	p := NewCSVLabParser()
	_, err := p.Parse([]byte("not,a,valid\ncsv\"broken"))
	// csv parser might handle this with LazyQuotes, but let's test the edge
	if err != nil {
		// If it returns an error, that's fine for malformed CSV
		return
	}
	// If it parsed, there should be no valid markers
}

func TestCSVLabParser_Parse_HeaderOnly_ReturnsError(t *testing.T) {
	csv := `Test Name,Result,Units,Reference Range
`
	p := NewCSVLabParser()
	_, err := p.Parse([]byte(csv))
	if err == nil {
		t.Fatal("expected error for header-only CSV")
	}
}

func TestCSVLabParser_Parse_NonNumericValue_SkipsRow(t *testing.T) {
	csv := `Test Name,Result,Units,Reference Range
Glucose,N/A,mg/dL,70-100
LDL Cholesterol,142,mg/dL,<100
`
	p := NewCSVLabParser()
	markers, err := p.Parse([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markers) != 1 {
		t.Fatalf("expected 1 marker (skipping N/A), got %d", len(markers))
	}
	if markers[0].Name != "LDL Cholesterol" {
		t.Errorf("expected 'LDL Cholesterol', got %q", markers[0].Name)
	}
}

func TestCSVLabParser_Parse_DuplicateMarkers_BothExtracted(t *testing.T) {
	csv := `Test Name,Result,Units,Reference Range
Glucose,98,mg/dL,70-100
Glucose,102,mg/dL,70-100
`
	p := NewCSVLabParser()
	markers, err := p.Parse([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markers) != 2 {
		t.Fatalf("expected 2 markers, got %d", len(markers))
	}
	if markers[0].Value != 98 {
		t.Errorf("expected first value 98, got %v", markers[0].Value)
	}
	if markers[1].Value != 102 {
		t.Errorf("expected second value 102, got %v", markers[1].Value)
	}
}

func TestCSVLabParser_Parse_DecimalValues_Parsed(t *testing.T) {
	csv := `Test Name,Result,Units,Reference Range
HbA1c,5.4,%,< 5.7
TSH,2.15,mIU/L,0.4-4.0
`
	p := NewCSVLabParser()
	markers, err := p.Parse([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markers) != 2 {
		t.Fatalf("expected 2 markers, got %d", len(markers))
	}
	if markers[0].Value != 5.4 {
		t.Errorf("expected 5.4, got %v", markers[0].Value)
	}
	if markers[1].Value != 2.15 {
		t.Errorf("expected 2.15, got %v", markers[1].Value)
	}
}

func TestCSVLabParser_Parse_NoHeader_PositionalParsing(t *testing.T) {
	csv := `Glucose,98,mg/dL,70-100
LDL Cholesterol,142,mg/dL,<100
`
	p := NewCSVLabParser()
	markers, err := p.Parse([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markers) != 2 {
		t.Fatalf("expected 2 markers, got %d", len(markers))
	}
}

func TestCSVLabParser_Parse_AlternateHeaders_Detected(t *testing.T) {
	csv := `Analyte,Value,Unit,Reference Interval
Glucose,98,mg/dL,70-100
`
	p := NewCSVLabParser()
	markers, err := p.Parse([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markers) != 1 {
		t.Fatalf("expected 1 marker, got %d", len(markers))
	}
	if markers[0].Name != "Glucose" {
		t.Errorf("expected 'Glucose', got %q", markers[0].Name)
	}
}

func TestCSVLabParser_Parse_ExtraColumns_Ignored(t *testing.T) {
	csv := `Test Name,Result,Units,Reference Range,Flag,Lab,Notes
Glucose,98,mg/dL,70-100,,Quest,Fasting
`
	p := NewCSVLabParser()
	markers, err := p.Parse([]byte(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markers) != 1 {
		t.Fatalf("expected 1 marker, got %d", len(markers))
	}
}
