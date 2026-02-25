package parser

import (
	"os"
	"strings"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

const testClientID = "client-abc-123"

// ── WHOOP Parser ────────────────────────────────────────────────────────────

func TestWhoopParser_Source_ReturnsWhoop(t *testing.T) {
	p := NewWhoopParser()
	if p.Source() != entities.WearableSourceWhoop {
		t.Errorf("expected source %q, got %q", entities.WearableSourceWhoop, p.Source())
	}
}

func TestWhoopParser_DetectFormat_ValidHeader_ReturnsTrue(t *testing.T) {
	header := []byte("Cycle start time,Cycle end time,Heart Rate Variability (ms),Resting Heart Rate (bpm),Sleep Performance %,Recovery Score %,Strain")
	p := NewWhoopParser()
	if !p.DetectFormat(header) {
		t.Error("expected DetectFormat to return true for WHOOP header")
	}
}

func TestWhoopParser_DetectFormat_InvalidHeader_ReturnsFalse(t *testing.T) {
	header := []byte("Date,Steps,Resting Heart Rate")
	p := NewWhoopParser()
	if p.DetectFormat(header) {
		t.Error("expected DetectFormat to return false for non-WHOOP header")
	}
}

func TestWhoopParser_Parse_SampleFile_ExtractsCorrectMeasurements(t *testing.T) {
	f, err := os.Open("testdata/whoop_sample.csv")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	p := NewWhoopParser()
	ms, err := p.Parse(f, testClientID)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// 3 rows × 5 metrics = 15 measurements
	if len(ms) != 15 {
		t.Fatalf("expected 15 measurements, got %d", len(ms))
	}

	// Verify first row metrics
	assertMeasurement(t, ms[0], entities.MeasurementTypeHRV, 45.2)
	assertMeasurement(t, ms[1], entities.MeasurementTypeRestingHR, 58)
	assertMeasurement(t, ms[2], entities.MeasurementTypeSleepQuality, 85)
	assertMeasurement(t, ms[3], entities.MeasurementTypeRecovery, 72)
	assertMeasurement(t, ms[4], entities.MeasurementTypeStrain, 14.5)

	// All measurements should have the correct client ID and source
	for _, m := range ms {
		if m.ClientID != testClientID {
			t.Errorf("expected client ID %q, got %q", testClientID, m.ClientID)
		}
		if m.Source != entities.WearableSourceWhoop {
			t.Errorf("expected source %q, got %q", entities.WearableSourceWhoop, m.Source)
		}
	}
}

func TestWhoopParser_Parse_EmptyInput_ReturnsError(t *testing.T) {
	p := NewWhoopParser()
	_, err := p.Parse(strings.NewReader(""), testClientID)
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestWhoopParser_Parse_HeaderOnly_ReturnsError(t *testing.T) {
	p := NewWhoopParser()
	_, err := p.Parse(strings.NewReader("Cycle start time,Heart Rate Variability (ms),Resting Heart Rate (bpm),Sleep Performance %,Recovery Score %,Strain\n"), testClientID)
	if err == nil {
		t.Fatal("expected error for header-only input")
	}
}

func TestWhoopParser_Parse_MissingColumns_ExtractsAvailable(t *testing.T) {
	input := "Cycle start time,Heart Rate Variability (ms)\n2024-01-01 06:00:00,45.2\n"
	p := NewWhoopParser()
	ms, err := p.Parse(strings.NewReader(input), testClientID)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(ms) != 1 {
		t.Fatalf("expected 1 measurement, got %d", len(ms))
	}
	assertMeasurement(t, ms[0], entities.MeasurementTypeHRV, 45.2)
}

// ── Garmin Parser ───────────────────────────────────────────────────────────

func TestGarminParser_Source_ReturnsGarmin(t *testing.T) {
	p := NewGarminParser()
	if p.Source() != entities.WearableSourceGarmin {
		t.Errorf("expected source %q, got %q", entities.WearableSourceGarmin, p.Source())
	}
}

func TestGarminParser_DetectFormat_ValidHeader_ReturnsTrue(t *testing.T) {
	header := []byte("Date,Steps,Resting Heart Rate (bpm),Sleep Duration (hrs),Stress Level,Heart Rate Variability (ms)")
	p := NewGarminParser()
	if !p.DetectFormat(header) {
		t.Error("expected DetectFormat to return true for Garmin header")
	}
}

func TestGarminParser_DetectFormat_InvalidHeader_ReturnsFalse(t *testing.T) {
	header := []byte("Recovery Score %,Strain")
	p := NewGarminParser()
	if p.DetectFormat(header) {
		t.Error("expected DetectFormat to return false for non-Garmin header")
	}
}

func TestGarminParser_Parse_SampleFile_ExtractsCorrectMeasurements(t *testing.T) {
	f, err := os.Open("testdata/garmin_sample.csv")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	p := NewGarminParser()
	ms, err := p.Parse(f, testClientID)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// 3 rows × 4 metrics = 12 measurements
	if len(ms) != 12 {
		t.Fatalf("expected 12 measurements, got %d", len(ms))
	}

	assertMeasurement(t, ms[0], entities.MeasurementTypeSteps, 8500)
	assertMeasurement(t, ms[1], entities.MeasurementTypeRestingHR, 62)
	assertMeasurement(t, ms[2], entities.MeasurementTypeSleepDuration, 7.5)
	assertMeasurement(t, ms[3], entities.MeasurementTypeHRV, 42.0)
}

func TestGarminParser_Parse_EmptyInput_ReturnsError(t *testing.T) {
	p := NewGarminParser()
	_, err := p.Parse(strings.NewReader(""), testClientID)
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

// ── Apple Health Parser ─────────────────────────────────────────────────────

func TestAppleHealthParser_Source_ReturnsAppleHealth(t *testing.T) {
	p := NewAppleHealthParser()
	if p.Source() != entities.WearableSourceAppleHealth {
		t.Errorf("expected source %q, got %q", entities.WearableSourceAppleHealth, p.Source())
	}
}

func TestAppleHealthParser_DetectFormat_ValidHeader_ReturnsTrue(t *testing.T) {
	header := []byte("Date,Heart Rate Variability (ms),Resting Heart Rate (bpm),Sleep Duration (hrs),Steps")
	p := NewAppleHealthParser()
	if !p.DetectFormat(header) {
		t.Error("expected DetectFormat to return true for Apple Health header")
	}
}

func TestAppleHealthParser_DetectFormat_WhoopHeader_ReturnsFalse(t *testing.T) {
	header := []byte("Cycle start time,Heart Rate Variability (ms),Recovery Score %,Strain")
	p := NewAppleHealthParser()
	if p.DetectFormat(header) {
		t.Error("expected DetectFormat to return false for WHOOP header")
	}
}

func TestAppleHealthParser_Parse_SampleFile_ExtractsCorrectMeasurements(t *testing.T) {
	f, err := os.Open("testdata/apple_health_sample.csv")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	p := NewAppleHealthParser()
	ms, err := p.Parse(f, testClientID)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// 3 rows × 4 metrics = 12 measurements
	if len(ms) != 12 {
		t.Fatalf("expected 12 measurements, got %d", len(ms))
	}

	assertMeasurement(t, ms[0], entities.MeasurementTypeHRV, 50.3)
	assertMeasurement(t, ms[1], entities.MeasurementTypeRestingHR, 55)
	assertMeasurement(t, ms[2], entities.MeasurementTypeSleepDuration, 7.8)
	assertMeasurement(t, ms[3], entities.MeasurementTypeSteps, 9200)
}

func TestAppleHealthParser_Parse_EmptyInput_ReturnsError(t *testing.T) {
	p := NewAppleHealthParser()
	_, err := p.Parse(strings.NewReader(""), testClientID)
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

// ── Oura Parser ─────────────────────────────────────────────────────────────

func TestOuraParser_Source_ReturnsOura(t *testing.T) {
	p := NewOuraParser()
	if p.Source() != entities.WearableSourceOura {
		t.Errorf("expected source %q, got %q", entities.WearableSourceOura, p.Source())
	}
}

func TestOuraParser_DetectFormat_ValidHeader_ReturnsTrue(t *testing.T) {
	header := []byte(`[{"date":"2024-01-01","hrv_ms":55,"recovery_score":82}]`)
	p := NewOuraParser()
	if !p.DetectFormat(header) {
		t.Error("expected DetectFormat to return true for Oura JSON")
	}
}

func TestOuraParser_DetectFormat_FitbitHeader_ReturnsFalse(t *testing.T) {
	header := []byte(`[{"date":"2024-01-01","hrv_ms":40,"steps":11000}]`)
	p := NewOuraParser()
	if p.DetectFormat(header) {
		t.Error("expected DetectFormat to return false for Fitbit JSON (no recovery_score)")
	}
}

func TestOuraParser_Parse_SampleFile_ExtractsCorrectMeasurements(t *testing.T) {
	f, err := os.Open("testdata/oura_sample.json")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	p := NewOuraParser()
	ms, err := p.Parse(f, testClientID)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// 3 rows × 5 metrics = 15 measurements
	if len(ms) != 15 {
		t.Fatalf("expected 15 measurements, got %d", len(ms))
	}

	assertMeasurement(t, ms[0], entities.MeasurementTypeHRV, 55.0)
	assertMeasurement(t, ms[1], entities.MeasurementTypeRestingHR, 52)
	assertMeasurement(t, ms[2], entities.MeasurementTypeSleepDuration, 7.9)
	assertMeasurement(t, ms[3], entities.MeasurementTypeRecovery, 82)
	assertMeasurement(t, ms[4], entities.MeasurementTypeSteps, 7500)
}

func TestOuraParser_Parse_EmptyArray_ReturnsError(t *testing.T) {
	p := NewOuraParser()
	_, err := p.Parse(strings.NewReader("[]"), testClientID)
	if err == nil {
		t.Fatal("expected error for empty array")
	}
}

func TestOuraParser_Parse_InvalidJSON_ReturnsError(t *testing.T) {
	p := NewOuraParser()
	_, err := p.Parse(strings.NewReader("{bad json"), testClientID)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// ── Fitbit Parser ───────────────────────────────────────────────────────────

func TestFitbitParser_Source_ReturnsFitbit(t *testing.T) {
	p := NewFitbitParser()
	if p.Source() != entities.WearableSourceFitbit {
		t.Errorf("expected source %q, got %q", entities.WearableSourceFitbit, p.Source())
	}
}

func TestFitbitParser_DetectFormat_ValidHeader_ReturnsTrue(t *testing.T) {
	header := []byte(`[{"date":"2024-01-01","hrv_ms":40,"steps":11000}]`)
	p := NewFitbitParser()
	if !p.DetectFormat(header) {
		t.Error("expected DetectFormat to return true for Fitbit JSON")
	}
}

func TestFitbitParser_DetectFormat_OuraHeader_ReturnsFalse(t *testing.T) {
	header := []byte(`[{"date":"2024-01-01","hrv_ms":55,"recovery_score":82}]`)
	p := NewFitbitParser()
	if p.DetectFormat(header) {
		t.Error("expected DetectFormat to return false for Oura JSON (has recovery_score)")
	}
}

func TestFitbitParser_Parse_SampleFile_ExtractsCorrectMeasurements(t *testing.T) {
	f, err := os.Open("testdata/fitbit_sample.json")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	p := NewFitbitParser()
	ms, err := p.Parse(f, testClientID)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// 3 rows × 4 metrics = 12 measurements
	if len(ms) != 12 {
		t.Fatalf("expected 12 measurements, got %d", len(ms))
	}

	assertMeasurement(t, ms[0], entities.MeasurementTypeHRV, 40.5)
	assertMeasurement(t, ms[1], entities.MeasurementTypeRestingHR, 65)
	assertMeasurement(t, ms[2], entities.MeasurementTypeSleepDuration, 6.8)
	assertMeasurement(t, ms[3], entities.MeasurementTypeSteps, 11000)
}

func TestFitbitParser_Parse_EmptyArray_ReturnsError(t *testing.T) {
	p := NewFitbitParser()
	_, err := p.Parse(strings.NewReader("[]"), testClientID)
	if err == nil {
		t.Fatal("expected error for empty array")
	}
}

func TestFitbitParser_Parse_InvalidJSON_ReturnsError(t *testing.T) {
	p := NewFitbitParser()
	_, err := p.Parse(strings.NewReader("not json"), testClientID)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// ── Registry ────────────────────────────────────────────────────────────────

func TestRegistry_Detect_WhoopCSV_ReturnsWhoopParser(t *testing.T) {
	reg := NewRegistry()
	header := []byte("Cycle start time,Recovery Score %,Strain,Heart Rate Variability (ms)")
	p := reg.Detect(header)
	if p == nil {
		t.Fatal("expected parser, got nil")
	}
	if p.Source() != entities.WearableSourceWhoop {
		t.Errorf("expected whoop, got %s", p.Source())
	}
}

func TestRegistry_Detect_GarminCSV_ReturnsGarminParser(t *testing.T) {
	reg := NewRegistry()
	header := []byte("Date,Steps,Stress Level")
	p := reg.Detect(header)
	if p == nil {
		t.Fatal("expected parser, got nil")
	}
	if p.Source() != entities.WearableSourceGarmin {
		t.Errorf("expected garmin, got %s", p.Source())
	}
}

func TestRegistry_Detect_OuraJSON_ReturnsOuraParser(t *testing.T) {
	reg := NewRegistry()
	header := []byte(`[{"date":"2024-01-01","hrv_ms":55,"recovery_score":82}]`)
	p := reg.Detect(header)
	if p == nil {
		t.Fatal("expected parser, got nil")
	}
	if p.Source() != entities.WearableSourceOura {
		t.Errorf("expected oura, got %s", p.Source())
	}
}

func TestRegistry_Detect_FitbitJSON_ReturnsFitbitParser(t *testing.T) {
	reg := NewRegistry()
	header := []byte(`[{"date":"2024-01-01","hrv_ms":40,"steps":11000}]`)
	p := reg.Detect(header)
	if p == nil {
		t.Fatal("expected parser, got nil")
	}
	if p.Source() != entities.WearableSourceFitbit {
		t.Errorf("expected fitbit, got %s", p.Source())
	}
}

func TestRegistry_Detect_UnknownFormat_ReturnsNil(t *testing.T) {
	reg := NewRegistry()
	header := []byte("completely unknown format")
	p := reg.Detect(header)
	if p != nil {
		t.Errorf("expected nil, got %s", p.Source())
	}
}

func TestRegistry_ForSource_KnownSource_ReturnsParser(t *testing.T) {
	reg := NewRegistry()
	sources := []entities.WearableSource{
		entities.WearableSourceWhoop,
		entities.WearableSourceGarmin,
		entities.WearableSourceAppleHealth,
		entities.WearableSourceOura,
		entities.WearableSourceFitbit,
	}
	for _, src := range sources {
		p := reg.ForSource(src)
		if p == nil {
			t.Errorf("expected parser for %s, got nil", src)
			continue
		}
		if p.Source() != src {
			t.Errorf("expected source %s, got %s", src, p.Source())
		}
	}
}

func TestRegistry_ForSource_UnknownSource_ReturnsNil(t *testing.T) {
	reg := NewRegistry()
	p := reg.ForSource("unknown")
	if p != nil {
		t.Errorf("expected nil for unknown source, got %s", p.Source())
	}
}

// ── All parsers implement WearableParser interface ──────────────────────────

func TestAllParsers_ImplementWearableParserInterface(t *testing.T) {
	parsers := []entities.WearableParser{
		NewWhoopParser(),
		NewGarminParser(),
		NewAppleHealthParser(),
		NewOuraParser(),
		NewFitbitParser(),
	}
	for _, p := range parsers {
		if p.Source() == "" {
			t.Error("parser returned empty source")
		}
	}
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func assertMeasurement(t *testing.T, m entities.Measurement, expectedType entities.MeasurementType, expectedValue float64) {
	t.Helper()
	if m.MeasurementType != expectedType {
		t.Errorf("expected type %q, got %q", expectedType, m.MeasurementType)
	}
	if m.Value != expectedValue {
		t.Errorf("expected value %f, got %f", expectedValue, m.Value)
	}
}
