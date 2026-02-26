package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// OuraParser parses Oura Ring daily summary CSV exports.
type OuraParser struct{}

// NewOuraParser creates a new Oura parser.
func NewOuraParser() *OuraParser {
	return &OuraParser{}
}

// Source returns the wearable source identifier.
func (p *OuraParser) Source() entities.WearableSource {
	return entities.WearableSourceOura
}

// DetectFormat returns true if the header matches Oura CSV format.
// Oura has HRV, Recovery Score, and Steps but no Strain Score.
func (p *OuraParser) DetectFormat(header []byte) bool {
	h := strings.ToLower(string(header))
	hasHRV := strings.Contains(h, "hrv (ms)")
	hasRecovery := strings.Contains(h, "recovery score")
	hasSteps := strings.Contains(h, "steps")
	hasNoStrain := !strings.Contains(h, "strain")
	return hasHRV && hasRecovery && hasSteps && hasNoStrain
}

// Parse reads an Oura CSV export and returns normalized measurements.
func (p *OuraParser) Parse(reader io.Reader, clientID, recordedBy string) ([]entities.Measurement, error) {
	r := csv.NewReader(reader)
	headers, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	colDate := csvColumnIndex(headers, "date")
	if colDate < 0 {
		return nil, fmt.Errorf("missing required column: Date")
	}

	colHRV := csvColumnIndex(headers, "hrv (ms)")
	colSleep := csvColumnIndex(headers, "sleep duration (hrs)")
	colRecovery := csvColumnIndex(headers, "recovery score")
	colSteps := csvColumnIndex(headers, "steps")
	colRestingHR := csvColumnIndex(headers, "resting hr (bpm)")

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

		measurements = appendMeasurement(measurements, row, colHRV, entities.MeasurementTypeHRV, clientID, recordedBy, measuredAt, entities.WearableSourceOura)
		measurements = appendMeasurement(measurements, row, colSleep, entities.MeasurementTypeSleepDuration, clientID, recordedBy, measuredAt, entities.WearableSourceOura)
		measurements = appendMeasurement(measurements, row, colRecovery, entities.MeasurementTypeRecovery, clientID, recordedBy, measuredAt, entities.WearableSourceOura)
		measurements = appendMeasurement(measurements, row, colSteps, entities.MeasurementTypeSteps, clientID, recordedBy, measuredAt, entities.WearableSourceOura)
		measurements = appendMeasurement(measurements, row, colRestingHR, entities.MeasurementTypeRestingHR, clientID, recordedBy, measuredAt, entities.WearableSourceOura)
	}

	if len(measurements) == 0 {
		return nil, fmt.Errorf("no valid measurements found in Oura export")
	}

	return measurements, nil
}
