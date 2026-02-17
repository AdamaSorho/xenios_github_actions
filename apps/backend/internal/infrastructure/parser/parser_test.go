package parser

import (
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
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
	header := []byte("Cycle Start Time,Cycle End Time,HRV (ms),Resting Heart Rate (bpm),Recovery Score (%),Sleep Duration (hrs),Strain Score")
	if !p.DetectFormat(header) {
		t.Error("expected DetectFormat to return true for valid WHOOP header")
	}
}

func TestWhoopParser_DetectFormat_InvalidHeader_ReturnsFalse(t *testing.T) {
	p := NewWhoopParser()
	header := []byte("Date,Steps,Resting Heart Rate (bpm)")
	if p.DetectFormat(header) {
		t.Error("expected DetectFormat to return false for non-WHOOP header")
	}
}

func TestWhoopParser_Parse_ValidCSV_ReturnsMeasurements(t *testing.T) {
	p := NewWhoopParser()
	f, err := os.Open("testdata/whoop_export.csv")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	measurements, err := p.Parse(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 3 rows * 5 metrics (HRV, sleep, recovery, strain, resting HR) = 15
	if len(measurements) != 15 {
		t.Fatalf("expected 15 measurements, got %d", len(measurements))
	}

	// Verify first row HRV
	found := false
	for _, m := range measurements {
		if m.MeasurementType == entities.MeasurementTypeHRV &&
			m.MeasuredAt.Day() == 1 {
			found = true
			if !floatEquals(m.Value, 45.2) {
				t.Errorf("expected HRV 45.2, got %f", m.Value)
			}
			if m.Unit != "ms" {
				t.Errorf("expected unit 'ms', got %q", m.Unit)
			}
			if m.Source != entities.WearableSourceWhoop {
				t.Errorf("expected source whoop, got %q", m.Source)
			}
		}
	}
	if !found {
		t.Error("expected to find HRV measurement for Jan 1")
	}
}

func TestWhoopParser_Parse_PartialColumns_ExtractsAvailable(t *testing.T) {
	p := NewWhoopParser()
	f, err := os.Open("testdata/whoop_partial.csv")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	measurements, err := p.Parse(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 2 rows * 2 metrics (HRV, recovery) = 4
	if len(measurements) != 4 {
		t.Fatalf("expected 4 measurements, got %d", len(measurements))
	}
}

func TestWhoopParser_Parse_EmptyReader_ReturnsEmpty(t *testing.T) {
	p := NewWhoopParser()
	measurements, err := p.Parse(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for empty reader")
	}
	if measurements != nil {
		t.Errorf("expected nil measurements, got %d", len(measurements))
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
	header := []byte("Date,Steps,Resting Heart Rate (bpm),HRV (ms),Sleep Duration (hrs),Stress Score")
	if !p.DetectFormat(header) {
		t.Error("expected DetectFormat to return true for valid Garmin header")
	}
}

func TestGarminParser_DetectFormat_InvalidHeader_ReturnsFalse(t *testing.T) {
	p := NewGarminParser()
	header := []byte("Cycle Start Time,HRV (ms)")
	if p.DetectFormat(header) {
		t.Error("expected DetectFormat to return false for non-Garmin header")
	}
}

func TestGarminParser_Parse_ValidCSV_ReturnsMeasurements(t *testing.T) {
	p := NewGarminParser()
	f, err := os.Open("testdata/garmin_export.csv")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	measurements, err := p.Parse(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 3 rows * 4 metrics (HRV, sleep, steps, resting HR) = 12
	if len(measurements) != 12 {
		t.Fatalf("expected 12 measurements, got %d", len(measurements))
	}

	// Verify first row steps
	found := false
	for _, m := range measurements {
		if m.MeasurementType == entities.MeasurementTypeSteps &&
			m.MeasuredAt.Day() == 1 {
			found = true
			if !floatEquals(m.Value, 8500) {
				t.Errorf("expected steps 8500, got %f", m.Value)
			}
		}
	}
	if !found {
		t.Error("expected to find Steps measurement for Jan 1")
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
	header := []byte("Date,Heart Rate Variability (ms),Resting Heart Rate (bpm),Sleep Duration (hrs),Steps")
	if !p.DetectFormat(header) {
		t.Error("expected DetectFormat to return true for valid Apple Health header")
	}
}

func TestAppleHealthParser_Parse_ValidCSV_ReturnsMeasurements(t *testing.T) {
	p := NewAppleHealthParser()
	f, err := os.Open("testdata/apple_health_export.csv")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	measurements, err := p.Parse(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 3 rows * 4 metrics (HRV, sleep, steps, resting HR) = 12
	if len(measurements) != 12 {
		t.Fatalf("expected 12 measurements, got %d", len(measurements))
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
	header := []byte("Date,HRV Average (ms),Resting Heart Rate (bpm),Recovery Index,Sleep Duration (hrs),Steps")
	if !p.DetectFormat(header) {
		t.Error("expected DetectFormat to return true for valid Oura header")
	}
}

func TestOuraParser_Parse_ValidCSV_ReturnsMeasurements(t *testing.T) {
	p := NewOuraParser()
	f, err := os.Open("testdata/oura_export.csv")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	measurements, err := p.Parse(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 3 rows * 5 metrics (HRV, sleep, recovery, steps, resting HR) = 15
	if len(measurements) != 15 {
		t.Fatalf("expected 15 measurements, got %d", len(measurements))
	}
}

// --- Fitbit Parser Tests ---

func TestFitbitParser_Source_ReturnsFitbit(t *testing.T) {
	p := NewFitbitParser()
	if p.Source() != entities.WearableSourceFitbit {
		t.Errorf("expected source %q, got %q", entities.WearableSourceFitbit, p.Source())
	}
}

func TestFitbitParser_DetectFormat_ValidHeader_ReturnsTrue(t *testing.T) {
	p := NewFitbitParser()
	header := []byte("Date,Resting Heart Rate,HRV (ms),Sleep Duration (hrs),Steps")
	if !p.DetectFormat(header) {
		t.Error("expected DetectFormat to return true for valid Fitbit header")
	}
}

func TestFitbitParser_Parse_ValidCSV_ReturnsMeasurements(t *testing.T) {
	p := NewFitbitParser()
	f, err := os.Open("testdata/fitbit_export.csv")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	measurements, err := p.Parse(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 3 rows * 4 metrics (HRV, sleep, steps, resting HR) = 12
	if len(measurements) != 12 {
		t.Fatalf("expected 12 measurements, got %d", len(measurements))
	}
}

// --- DetectSource Tests ---

func TestDetectSource_WhoopHeader_ReturnsWhoopParser(t *testing.T) {
	header := []byte("Cycle Start Time,Cycle End Time,HRV (ms),Resting Heart Rate (bpm),Recovery Score (%),Sleep Duration (hrs),Strain Score\n")
	p, err := DetectSource(header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Source() != entities.WearableSourceWhoop {
		t.Errorf("expected whoop source, got %q", p.Source())
	}
}

func TestDetectSource_GarminHeader_ReturnsGarminParser(t *testing.T) {
	header := []byte("Date,Steps,Resting Heart Rate (bpm),HRV (ms),Sleep Duration (hrs),Stress Score\n")
	p, err := DetectSource(header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Source() != entities.WearableSourceGarmin {
		t.Errorf("expected garmin source, got %q", p.Source())
	}
}

func TestDetectSource_UnknownHeader_ReturnsError(t *testing.T) {
	header := []byte("This,Is,Not,A,Valid,Wearable,Export\n")
	_, err := DetectSource(header)
	if err == nil {
		t.Fatal("expected error for unknown header format")
	}
}

// --- Malformed Data Tests ---

func TestWhoopParser_Parse_MalformedCSV_ReturnsError(t *testing.T) {
	p := NewWhoopParser()
	f, err := os.Open("testdata/malformed.csv")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	measurements, err := p.Parse(f)
	// Malformed: should return empty measurements (no recognizable columns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(measurements) != 0 {
		t.Errorf("expected 0 measurements from malformed data, got %d", len(measurements))
	}
}

// --- Measurement validation ---

func TestAllParsers_MeasuredAt_HasCorrectDate(t *testing.T) {
	parsers := []struct {
		name    string
		parser  WearableParser
		fixture string
	}{
		{"whoop", NewWhoopParser(), "testdata/whoop_export.csv"},
		{"garmin", NewGarminParser(), "testdata/garmin_export.csv"},
		{"apple_health", NewAppleHealthParser(), "testdata/apple_health_export.csv"},
		{"oura", NewOuraParser(), "testdata/oura_export.csv"},
		{"fitbit", NewFitbitParser(), "testdata/fitbit_export.csv"},
	}

	expectedDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for _, tt := range parsers {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.fixture)
			if err != nil {
				t.Fatalf("failed to open fixture: %v", err)
			}
			defer f.Close()

			measurements, err := tt.parser.Parse(f)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(measurements) == 0 {
				t.Fatal("expected at least one measurement")
			}

			// Find a measurement from the first date
			found := false
			for _, m := range measurements {
				if m.MeasuredAt.Year() == expectedDate.Year() &&
					m.MeasuredAt.Month() == expectedDate.Month() &&
					m.MeasuredAt.Day() == expectedDate.Day() {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected to find measurement for date %s", expectedDate.Format("2006-01-02"))
			}
		})
	}
}

func TestAllParsers_Source_MatchesExpected(t *testing.T) {
	parsers := []struct {
		parser   WearableParser
		expected entities.WearableSource
		fixture  string
	}{
		{NewWhoopParser(), entities.WearableSourceWhoop, "testdata/whoop_export.csv"},
		{NewGarminParser(), entities.WearableSourceGarmin, "testdata/garmin_export.csv"},
		{NewAppleHealthParser(), entities.WearableSourceAppleHealth, "testdata/apple_health_export.csv"},
		{NewOuraParser(), entities.WearableSourceOura, "testdata/oura_export.csv"},
		{NewFitbitParser(), entities.WearableSourceFitbit, "testdata/fitbit_export.csv"},
	}

	for _, tt := range parsers {
		t.Run(string(tt.expected), func(t *testing.T) {
			f, err := os.Open(tt.fixture)
			if err != nil {
				t.Fatalf("failed to open fixture: %v", err)
			}
			defer f.Close()

			measurements, err := tt.parser.Parse(f)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, m := range measurements {
				if m.Source != tt.expected {
					t.Errorf("expected source %q, got %q", tt.expected, m.Source)
				}
			}
		})
	}
}

// floatEquals compares two floats with a small tolerance.
func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < 0.001
}
