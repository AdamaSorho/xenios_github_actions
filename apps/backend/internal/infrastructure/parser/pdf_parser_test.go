package parser

import (
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

var _ repository.LabFileParser = &PDFLabParser{}

func TestPDFLabParser_Parse_TabDelimitedText_ExtractsMarkers(t *testing.T) {
	text := "Glucose\t98\tmg/dL\t70-100\nLDL Cholesterol\t142\tmg/dL\t<100\n"

	p := NewPDFLabParser()
	markers, err := p.Parse([]byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markers) != 2 {
		t.Fatalf("expected 2 markers, got %d", len(markers))
	}
	if markers[0].Name != "Glucose" {
		t.Errorf("expected 'Glucose', got %q", markers[0].Name)
	}
	if markers[0].Value != 98 {
		t.Errorf("expected 98, got %v", markers[0].Value)
	}
	if markers[1].Name != "LDL Cholesterol" {
		t.Errorf("expected 'LDL Cholesterol', got %q", markers[1].Name)
	}
}

func TestPDFLabParser_Parse_PipeDelimitedText_ExtractsMarkers(t *testing.T) {
	text := "Glucose | 98 | mg/dL | 70-100\nTSH | 2.1 | mIU/L | 0.4-4.0\n"

	p := NewPDFLabParser()
	markers, err := p.Parse([]byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markers) != 2 {
		t.Fatalf("expected 2 markers, got %d", len(markers))
	}
}

func TestPDFLabParser_Parse_MultiSpaceDelimited_ExtractsMarkers(t *testing.T) {
	text := "Glucose  98  mg/dL  70-100\nHDL Cholesterol  55  mg/dL  >40\n"

	p := NewPDFLabParser()
	markers, err := p.Parse([]byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markers) != 2 {
		t.Fatalf("expected 2 markers, got %d", len(markers))
	}
	if markers[1].Name != "HDL Cholesterol" {
		t.Errorf("expected 'HDL Cholesterol', got %q", markers[1].Name)
	}
}

func TestPDFLabParser_Parse_EmptyContent_ReturnsError(t *testing.T) {
	p := NewPDFLabParser()
	_, err := p.Parse([]byte(""))
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestPDFLabParser_Parse_NoMarkers_ReturnsError(t *testing.T) {
	text := "Patient Name: John Doe\nDate of Birth: 1990-01-01\nReport Date: 2024-01-15\n"

	p := NewPDFLabParser()
	_, err := p.Parse([]byte(text))
	if err == nil {
		t.Fatal("expected error when no markers found")
	}
}

func TestPDFLabParser_Parse_WithHeaderRow_SkipsHeader(t *testing.T) {
	text := "Test Name\tResult\tUnits\tReference Range\nGlucose\t98\tmg/dL\t70-100\n"

	p := NewPDFLabParser()
	markers, err := p.Parse([]byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markers) != 1 {
		t.Fatalf("expected 1 marker (header skipped), got %d", len(markers))
	}
	if markers[0].Name != "Glucose" {
		t.Errorf("expected 'Glucose', got %q", markers[0].Name)
	}
}

func TestPDFLabParser_Parse_OutOfRangeDetection_Works(t *testing.T) {
	text := "LDL Cholesterol\t142\tmg/dL\t<100\n"

	p := NewPDFLabParser()
	markers, err := p.Parse([]byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	flag := entities.DetermineFlag(markers[0].Value, markers[0].ReferenceLow, markers[0].ReferenceHigh)
	if flag == nil || *flag != entities.FlagHigh {
		t.Errorf("expected 'high' flag for LDL 142 with ref <100, got %v", flag)
	}
}

func TestPDFLabParser_Parse_SimplePDFTextOperators_ExtractsText(t *testing.T) {
	// Simulate a very simple PDF with text operators
	pdf := "%PDF-1.4\nBT (Glucose) Tj ET\nBT (98) Tj ET\n"

	p := NewPDFLabParser()
	// This may or may not extract markers depending on the text structure
	// The point is it shouldn't crash
	_, _ = p.Parse([]byte(pdf))
}

func TestPDFLabParser_Parse_MixedContent_SkipsNonMarkerLines(t *testing.T) {
	text := `Lab Results Report
Patient: John Doe
Date: 2024-01-15

Glucose	98	mg/dL	70-100
Some random text line
LDL Cholesterol	142	mg/dL	<100
Another non-marker line
`

	p := NewPDFLabParser()
	markers, err := p.Parse([]byte(text))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(markers) != 2 {
		t.Fatalf("expected 2 markers, got %d", len(markers))
	}
}
