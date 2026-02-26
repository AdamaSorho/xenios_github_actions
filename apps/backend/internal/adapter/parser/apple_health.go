package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// AppleHealthParser parses Apple Health CSV exports.
type AppleHealthParser struct{}

// NewAppleHealthParser creates a new Apple Health parser.
func NewAppleHealthParser() *AppleHealthParser {
	return &AppleHealthParser{}
}

// Source returns the wearable source identifier.
func (p *AppleHealthParser) Source() entities.WearableSource {
	return entities.WearableSourceAppleHealth
}

// DetectFormat returns true if the header matches Apple Health CSV format.
// Apple Health has Steps and HRV columns but lacks Recovery Score, Strain Score, and Stress Level.
func (p *AppleHealthParser) DetectFormat(header []byte) bool {
	h := strings.ToLower(string(header))
	hasSteps := strings.Contains(h, "steps")
	hasHRV := strings.Contains(h, "heart rate variability")
	hasNoRecovery := !strings.Contains(h, "recovery")
	hasNoStrain := !strings.Contains(h, "strain")
	hasNoStress := !strings.Contains(h, "stress")
	return hasSteps && hasHRV && hasNoRecovery && hasNoStrain && hasNoStress
}

// Parse reads an Apple Health CSV export and returns normalized measurements.
func (p *AppleHealthParser) Parse(reader io.Reader, clientID, recordedBy string) ([]entities.Measurement, error) {
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

		measurements = appendMeasurement(measurements, row, colSteps, entities.MeasurementTypeSteps, clientID, recordedBy, measuredAt, entities.WearableSourceAppleHealth)
		measurements = appendMeasurement(measurements, row, colSleep, entities.MeasurementTypeSleepDuration, clientID, recordedBy, measuredAt, entities.WearableSourceAppleHealth)
		measurements = appendMeasurement(measurements, row, colHRV, entities.MeasurementTypeHRV, clientID, recordedBy, measuredAt, entities.WearableSourceAppleHealth)
		measurements = appendMeasurement(measurements, row, colRestingHR, entities.MeasurementTypeRestingHR, clientID, recordedBy, measuredAt, entities.WearableSourceAppleHealth)
	}

	if len(measurements) == 0 {
		return nil, fmt.Errorf("no valid measurements found in Apple Health export")
	}

	return measurements, nil
}
