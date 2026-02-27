package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// GarminParser parses Garmin daily summary CSV exports.
type GarminParser struct{}

// NewGarminParser creates a new Garmin parser.
func NewGarminParser() *GarminParser {
	return &GarminParser{}
}

// Source returns the wearable source identifier.
func (p *GarminParser) Source() entities.WearableSource {
	return entities.WearableSourceGarmin
}

// DetectFormat returns true if the header matches Garmin CSV format.
func (p *GarminParser) DetectFormat(header []byte) bool {
	h := strings.ToLower(string(header))
	return strings.Contains(h, "stress level") && strings.Contains(h, "steps")
}

// Parse reads a Garmin CSV export and returns normalized measurements.
func (p *GarminParser) Parse(reader io.Reader, clientID, recordedBy string) ([]entities.Measurement, error) {
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

		measurements = appendMeasurement(measurements, row, colSteps, entities.MeasurementTypeSteps, clientID, recordedBy, measuredAt, entities.WearableSourceGarmin)
		measurements = appendMeasurement(measurements, row, colSleep, entities.MeasurementTypeSleepDuration, clientID, recordedBy, measuredAt, entities.WearableSourceGarmin)
		measurements = appendMeasurement(measurements, row, colHRV, entities.MeasurementTypeHRV, clientID, recordedBy, measuredAt, entities.WearableSourceGarmin)
		measurements = appendMeasurement(measurements, row, colRestingHR, entities.MeasurementTypeRestingHR, clientID, recordedBy, measuredAt, entities.WearableSourceGarmin)
	}

	if len(measurements) == 0 {
		return nil, fmt.Errorf("no valid measurements found in Garmin export")
	}

	return measurements, nil
}
