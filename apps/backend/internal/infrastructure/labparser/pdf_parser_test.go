package labparser

import (
	"os"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestParsePDFText_LabReport_ExtractsMarkers(t *testing.T) {
	data, err := os.ReadFile(testdataPath("lab_report.txt"))
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	results, err := ParsePDFText(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) < 5 {
		t.Fatalf("expected at least 5 results, got %d", len(results))
	}

	// Verify glucose
	glucose := findResult(results, string(entities.LabFastingGlucose))
	if glucose == nil {
		t.Fatal("glucose result not found")
	}
	if glucose.Value != 98 {
		t.Errorf("glucose value: expected 98, got %f", glucose.Value)
	}
	if glucose.Flag != entities.FlagNormal {
		t.Errorf("glucose flag: expected normal, got %s", glucose.Flag)
	}

	// Verify LDL flagged high
	ldl := findResult(results, string(entities.LabLDLCholesterol))
	if ldl == nil {
		t.Fatal("LDL result not found")
	}
	if ldl.Flag != entities.FlagHigh {
		t.Errorf("LDL flag: expected high, got %s", ldl.Flag)
	}
}

func TestParsePDFText_EmptyText_ReturnsEmpty(t *testing.T) {
	results, err := ParsePDFText([]byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty text, got %d", len(results))
	}
}

func TestParsePDFText_NoLabData_ReturnsEmpty(t *testing.T) {
	text := "This is just a regular document with no lab data in it."
	results, err := ParsePDFText([]byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for non-lab text, got %d", len(results))
	}
}

func TestParsePDFText_InlineResults_ExtractsMarkers(t *testing.T) {
	text := `Lab Results
Glucose  95 mg/dL  70-100
LDL Cholesterol  130 mg/dL  <100
TSH  3.2 mIU/L  0.4-4.0
`
	results, err := ParsePDFText([]byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	ldl := findResult(results, string(entities.LabLDLCholesterol))
	if ldl == nil {
		t.Fatal("LDL result not found")
	}
	if ldl.Value != 130 {
		t.Errorf("LDL value: expected 130, got %f", ldl.Value)
	}
	if ldl.Flag != entities.FlagHigh {
		t.Errorf("LDL flag: expected high, got %s", ldl.Flag)
	}
}

func TestDetectFormat_CSV(t *testing.T) {
	csvData := []byte("Test Name,Result,Units,Reference Range\nGlucose,95,mg/dL,70-100\n")
	format := DetectFormat(csvData, "results.csv")
	if format != FormatCSV {
		t.Errorf("expected CSV format, got %s", format)
	}
}

func TestDetectFormat_PDF(t *testing.T) {
	// PDF files start with %PDF magic bytes
	pdfData := []byte("%PDF-1.4 some binary content")
	format := DetectFormat(pdfData, "results.pdf")
	if format != FormatPDF {
		t.Errorf("expected PDF format, got %s", format)
	}
}

func TestDetectFormat_PDFByExtension(t *testing.T) {
	format := DetectFormat([]byte("some content"), "results.pdf")
	if format != FormatPDF {
		t.Errorf("expected PDF format by extension, got %s", format)
	}
}

func TestDetectFormat_CSVByExtension(t *testing.T) {
	format := DetectFormat([]byte("some content"), "results.csv")
	if format != FormatCSV {
		t.Errorf("expected CSV format by extension, got %s", format)
	}
}

func TestDetectFormat_Unknown(t *testing.T) {
	format := DetectFormat([]byte("random bytes"), "results.docx")
	if format != FormatUnknown {
		t.Errorf("expected unknown format, got %s", format)
	}
}
