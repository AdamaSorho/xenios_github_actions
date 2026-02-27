package parser

import (
	"os"
	"strings"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

const (
	testClientID   = "client-123"
	testRecordedBy = "coach-456"
)

// --- WHOOP Parser Tests ---

func TestWhoopParser_Source_ReturnsWhoop(t *testing.T) {
	p := NewWhoopParser()
	if p.Source() != entities.WearableSourceWhoop {
		t.Errorf("expected source %q, got %q", entities.WearableSourceWhoop, p.Source())
	}
}

func TestWhoopParser_DetectFormat_ValidHeader_ReturnsTrue(t *testing.T) {
	p := NewWhoopParser()
	header := []byte("Date,HRV (ms),Sleep Duration (hrs),Recovery Score,Strain Score,Resting HR (bpm)")
	if !p.DetectFormat(header) {
		t.Error("expected DetectFormat to return true for WHOOP header")
	}
}

func TestWhoopParser_DetectFormat_GarminHeader_ReturnsFalse(t *testing.T) {
	p := NewWhoopParser()
	header := []byte("Date,Steps,Sleep Duration (hrs),Heart Rate Variability (ms),Resting Heart Rate (bpm),Stress Level")
	if p.DetectFormat(header) {
		t.Error("expected DetectFormat to return false for Garmin header")
	}
}

func TestWhoopParser_Parse_ValidCSV_ReturnsMeasurements(t *testing.T) {
	p := NewWhoopParser()
	f, err := os.Open("testdata/whoop_export.csv")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	measurements, err := p.Parse(f, testClientID, testRecordedBy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 5 rows * 5 metrics = 25 measurements
	if len(measurements) != 25 {
		t.Errorf("expected 25 measurements, got %d", len(measurements))
	}

	// Verify first row's HRV
	found := false
	for _, m := range measurements {
		if m.MeasurementType == entities.MeasurementTypeHRV && m.MeasuredAt.Day() == 1 {
			found = true
			if m.Value != 45.2 {
				t.Errorf("expected HRV 45.2, got %f", m.Value)
			}
			if m.ClientID != testClientID {
				t.Errorf("expected clientID %q, got %q", testClientID, m.ClientID)
			}
			if m.Unit != "ms" {
				t.Errorf("expected unit %q, got %q", "ms", m.Unit)
			}
			if m.Source != entities.WearableSourceWhoop {
				t.Errorf("expected source %q, got %q", entities.WearableSourceWhoop, m.Source)
			}
			break
		}
	}
	if !found {
		t.Error("expected to find HRV measurement for Jan 1")
	}
}

func TestWhoopParser_Parse_MissingColumns_ExtractsAvailable(t *testing.T) {
	p := NewWhoopParser()
	f, err := os.Open("testdata/whoop_missing_columns.csv")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	measurements, err := p.Parse(f, testClientID, testRecordedBy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 2 rows * 2 available metrics (HRV, Recovery) = 4 measurements
	if len(measurements) != 4 {
		t.Errorf("expected 4 measurements, got %d", len(measurements))
	}
}

func TestWhoopParser_Parse_EmptyReader_ReturnsError(t *testing.T) {
	p := NewWhoopParser()
	_, err := p.Parse(strings.NewReader(""), testClientID, testRecordedBy)
	if err == nil {
		t.Error("expected error for empty reader")
	}
}

func TestWhoopParser_Parse_MissingDateColumn_ReturnsError(t *testing.T) {
	p := NewWhoopParser()
	csv := "HRV (ms),Recovery Score\n45.2,68\n"
	_, err := p.Parse(strings.NewReader(csv), testClientID, testRecordedBy)
	if err == nil {
		t.Error("expected error for missing Date column")
	}
}

func TestWhoopParser_Parse_InvalidDate_ReturnsError(t *testing.T) {
	p := NewWhoopParser()
	csv := "Date,HRV (ms)\nnot-a-date,45.2\n"
	_, err := p.Parse(strings.NewReader(csv), testClientID, testRecordedBy)
	if err == nil {
		t.Error("expected error for invalid date")
	}
}

func TestWhoopParser_Parse_InvalidValue_SkipsRow(t *testing.T) {
	p := NewWhoopParser()
	csv := "Date,HRV (ms),Recovery Score\n2024-01-01,not-a-number,68\n"
	measurements, err := p.Parse(strings.NewReader(csv), testClientID, testRecordedBy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only Recovery should be parsed, HRV has invalid value
	if len(measurements) != 1 {
		t.Errorf("expected 1 measurement, got %d", len(measurements))
	}
}

// --- Garmin Parser Tests ---

func TestGarminParser_Source_ReturnsGarmin(t *testing.T) {
	p := NewGarminParser()
	if p.Source() != entities.WearableSourceGarmin {
		t.Errorf("expected source %q, got %q", entities.WearableSourceGarmin, p.Source())
	}
}

func TestGarminParser_DetectFormat_ValidHeader_ReturnsTrue(t *testing.T) {
	p := NewGarminParser()
	header := []byte("Date,Steps,Sleep Duration (hrs),Heart Rate Variability (ms),Resting Heart Rate (bpm),Stress Level")
	if !p.DetectFormat(header) {
		t.Error("expected DetectFormat to return true for Garmin header")
	}
}

func TestGarminParser_DetectFormat_WhoopHeader_ReturnsFalse(t *testing.T) {
	p := NewGarminParser()
	header := []byte("Date,HRV (ms),Sleep Duration (hrs),Recovery Score,Strain Score,Resting HR (bpm)")
	if p.DetectFormat(header) {
		t.Error("expected DetectFormat to return false for WHOOP header")
	}
}

func TestGarminParser_Parse_ValidCSV_ReturnsMeasurements(t *testing.T) {
	p := NewGarminParser()
	f, err := os.Open("testdata/garmin_export.csv")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	measurements, err := p.Parse(f, testClientID, testRecordedBy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 5 rows * 4 metrics = 20 measurements
	if len(measurements) != 20 {
		t.Errorf("expected 20 measurements, got %d", len(measurements))
	}

	// Verify steps for first row
	found := false
	for _, m := range measurements {
		if m.MeasurementType == entities.MeasurementTypeSteps && m.MeasuredAt.Day() == 1 {
			found = true
			if m.Value != 8500 {
				t.Errorf("expected steps 8500, got %f", m.Value)
			}
			if m.Source != entities.WearableSourceGarmin {
				t.Errorf("expected source %q, got %q", entities.WearableSourceGarmin, m.Source)
			}
			break
		}
	}
	if !found {
		t.Error("expected to find steps measurement for Jan 1")
	}
}

func TestGarminParser_Parse_EmptyReader_ReturnsError(t *testing.T) {
	p := NewGarminParser()
	_, err := p.Parse(strings.NewReader(""), testClientID, testRecordedBy)
	if err == nil {
		t.Error("expected error for empty reader")
	}
}

// --- Apple Health Parser Tests ---

func TestAppleHealthParser_Source_ReturnsAppleHealth(t *testing.T) {
	p := NewAppleHealthParser()
	if p.Source() != entities.WearableSourceAppleHealth {
		t.Errorf("expected source %q, got %q", entities.WearableSourceAppleHealth, p.Source())
	}
}

func TestAppleHealthParser_DetectFormat_ValidHeader_ReturnsTrue(t *testing.T) {
	p := NewAppleHealthParser()
	header := []byte("Date,Steps,Sleep Duration (hrs),Heart Rate Variability (ms),Resting Heart Rate (bpm)")
	if !p.DetectFormat(header) {
		t.Error("expected DetectFormat to return true for Apple Health header")
	}
}

func TestAppleHealthParser_Parse_ValidCSV_ReturnsMeasurements(t *testing.T) {
	p := NewAppleHealthParser()
	f, err := os.Open("testdata/apple_health_export.csv")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	measurements, err := p.Parse(f, testClientID, testRecordedBy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 5 rows * 4 metrics = 20 measurements
	if len(measurements) != 20 {
		t.Errorf("expected 20 measurements, got %d", len(measurements))
	}

	// Verify source attribution
	for _, m := range measurements {
		if m.Source != entities.WearableSourceAppleHealth {
			t.Errorf("expected source %q, got %q", entities.WearableSourceAppleHealth, m.Source)
			break
		}
	}
}

// --- Oura Parser Tests ---

func TestOuraParser_Source_ReturnsOura(t *testing.T) {
	p := NewOuraParser()
	if p.Source() != entities.WearableSourceOura {
		t.Errorf("expected source %q, got %q", entities.WearableSourceOura, p.Source())
	}
}

func TestOuraParser_DetectFormat_ValidHeader_ReturnsTrue(t *testing.T) {
	p := NewOuraParser()
	header := []byte("Date,HRV (ms),Sleep Duration (hrs),Recovery Score,Steps,Resting HR (bpm)")
	if !p.DetectFormat(header) {
		t.Error("expected DetectFormat to return true for Oura header")
	}
}

func TestOuraParser_Parse_ValidCSV_ReturnsMeasurements(t *testing.T) {
	p := NewOuraParser()
	f, err := os.Open("testdata/oura_export.csv")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	measurements, err := p.Parse(f, testClientID, testRecordedBy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 5 rows * 5 metrics = 25 measurements
	if len(measurements) != 25 {
		t.Errorf("expected 25 measurements, got %d", len(measurements))
	}
}

// --- Fitbit Parser Tests ---

func TestFitbitParser_Source_ReturnsFitbit(t *testing.T) {
	p := NewFitbitParser()
	if p.Source() != entities.WearableSourceFitbit {
		t.Errorf("expected source %q, got %q", entities.WearableSourceFitbit, p.Source())
	}
}

func TestFitbitParser_Parse_ValidCSV_ReturnsMeasurements(t *testing.T) {
	p := NewFitbitParser()
	f, err := os.Open("testdata/fitbit_export.csv")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	measurements, err := p.Parse(f, testClientID, testRecordedBy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 5 rows * 4 metrics = 20 measurements
	if len(measurements) != 20 {
		t.Errorf("expected 20 measurements, got %d", len(measurements))
	}
}

// --- Registry Tests ---

func TestRegistry_DetectParser_WhoopHeader_ReturnsWhoopParser(t *testing.T) {
	reg := NewRegistry()
	header := []byte("Date,HRV (ms),Sleep Duration (hrs),Recovery Score,Strain Score,Resting HR (bpm)")
	p, err := reg.DetectParser(header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Source() != entities.WearableSourceWhoop {
		t.Errorf("expected source %q, got %q", entities.WearableSourceWhoop, p.Source())
	}
}

func TestRegistry_DetectParser_GarminHeader_ReturnsGarminParser(t *testing.T) {
	reg := NewRegistry()
	header := []byte("Date,Steps,Sleep Duration (hrs),Heart Rate Variability (ms),Resting Heart Rate (bpm),Stress Level")
	p, err := reg.DetectParser(header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Source() != entities.WearableSourceGarmin {
		t.Errorf("expected source %q, got %q", entities.WearableSourceGarmin, p.Source())
	}
}

func TestRegistry_DetectParser_UnknownHeader_ReturnsError(t *testing.T) {
	reg := NewRegistry()
	header := []byte("random,columns,that,dont,match")
	_, err := reg.DetectParser(header)
	if err == nil {
		t.Error("expected error for unknown header format")
	}
}

func TestRegistry_GetParser_ValidSource_ReturnsParser(t *testing.T) {
	reg := NewRegistry()
	sources := []entities.WearableSource{
		entities.WearableSourceWhoop,
		entities.WearableSourceGarmin,
		entities.WearableSourceAppleHealth,
		entities.WearableSourceOura,
		entities.WearableSourceFitbit,
	}

	for _, src := range sources {
		t.Run(string(src), func(t *testing.T) {
			p, err := reg.GetParser(src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Source() != src {
				t.Errorf("expected source %q, got %q", src, p.Source())
			}
		})
	}
}

func TestRegistry_GetParser_InvalidSource_ReturnsError(t *testing.T) {
	reg := NewRegistry()
	_, err := reg.GetParser("nonexistent")
	if err == nil {
		t.Error("expected error for invalid source")
	}
}

// --- csvColumnIndex Tests ---

func TestCsvColumnIndex_ExistingColumn_ReturnsIndex(t *testing.T) {
	headers := []string{"Date", "HRV (ms)", "Sleep Duration (hrs)"}
	idx := csvColumnIndex(headers, "hrv (ms)")
	if idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}
}

func TestCsvColumnIndex_CaseInsensitive_ReturnsIndex(t *testing.T) {
	headers := []string{"Date", "HRV (MS)", "Sleep Duration (hrs)"}
	idx := csvColumnIndex(headers, "hrv (ms)")
	if idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}
}

func TestCsvColumnIndex_MissingColumn_ReturnsNegative(t *testing.T) {
	headers := []string{"Date", "HRV (ms)"}
	idx := csvColumnIndex(headers, "nonexistent")
	if idx != -1 {
		t.Errorf("expected -1, got %d", idx)
	}
}

func TestCsvColumnIndex_TrimSpaces_ReturnsIndex(t *testing.T) {
	headers := []string{" Date ", " HRV (ms) "}
	idx := csvColumnIndex(headers, "hrv (ms)")
	if idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}
}
