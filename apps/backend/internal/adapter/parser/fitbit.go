package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// FitbitParser parses Fitbit daily summary CSV exports.
type FitbitParser struct{}

// NewFitbitParser creates a new Fitbit parser.
func NewFitbitParser() *FitbitParser {
	return &FitbitParser{}
}

// Source returns the wearable source identifier.
func (p *FitbitParser) Source() entities.WearableSource {
	return entities.WearableSourceFitbit
}

// DetectFormat returns true if the header matches Fitbit CSV format.
// Fitbit has Steps and HRV columns, no Recovery/Strain/Stress, and uses "Heart Rate Variability" header.
// This is similar to Apple Health, but Fitbit is checked after Apple Health in the registry.
// In practice the explicit source field on the job payload is used, so auto-detect is best-effort.
func (p *FitbitParser) DetectFormat(header []byte) bool {
	// Fitbit detection relies on explicit source selection since it overlaps with Apple Health.
	// Return false for auto-detection; use GetParser(source) instead.
	return false
}

// Parse reads a Fitbit CSV export and returns normalized measurements.
func (p *FitbitParser) Parse(reader io.Reader, clientID, recordedBy string) ([]entities.Measurement, error) {
	r := csv.NewReader(reader)
	headers, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	colDate := csvColumnIndex(headers, "date")
	if colDate < 0 {
		return nil, fmt.Errorf("missing required column: Date")
	}

	colSteps := csvColumnIndex(headers, "steps")
	colSleep := csvColumnIndex(headers, "sleep duration (hrs)")
	colHRV := csvColumnIndex(headers, "heart rate variability (ms)")
	colRestingHR := csvColumnIndex(headers, "resting heart rate (bpm)")

	var measurements []entities.Measurement
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		dateStr := strings.TrimSpace(row[colDate])
		measuredAt, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date %q: %w", dateStr, err)
		}

		measurements = appendMeasurement(measurements, row, colSteps, entities.MeasurementTypeSteps, clientID, recordedBy, measuredAt, entities.WearableSourceFitbit)
		measurements = appendMeasurement(measurements, row, colSleep, entities.MeasurementTypeSleepDuration, clientID, recordedBy, measuredAt, entities.WearableSourceFitbit)
		measurements = appendMeasurement(measurements, row, colHRV, entities.MeasurementTypeHRV, clientID, recordedBy, measuredAt, entities.WearableSourceFitbit)
		measurements = appendMeasurement(measurements, row, colRestingHR, entities.MeasurementTypeRestingHR, clientID, recordedBy, measuredAt, entities.WearableSourceFitbit)
	}

	if len(measurements) == 0 {
		return nil, fmt.Errorf("no valid measurements found in Fitbit export")
	}

	return measurements, nil
}
