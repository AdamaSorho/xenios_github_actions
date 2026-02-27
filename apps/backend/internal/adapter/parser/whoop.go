package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// WhoopParser parses WHOOP daily summary CSV exports.
type WhoopParser struct{}

// NewWhoopParser creates a new WHOOP parser.
func NewWhoopParser() *WhoopParser {
	return &WhoopParser{}
}

// Source returns the wearable source identifier.
func (p *WhoopParser) Source() entities.WearableSource {
	return entities.WearableSourceWhoop
}

// DetectFormat returns true if the header matches WHOOP CSV format.
func (p *WhoopParser) DetectFormat(header []byte) bool {
	h := strings.ToLower(string(header))
	return strings.Contains(h, "recovery score") && strings.Contains(h, "strain score")
}

// Parse reads a WHOOP CSV export and returns normalized measurements.
func (p *WhoopParser) Parse(reader io.Reader, clientID, recordedBy string) ([]entities.Measurement, error) {
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
	colStrain := csvColumnIndex(headers, "strain score")
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

		measurements = appendMeasurement(measurements, row, colHRV, entities.MeasurementTypeHRV, clientID, recordedBy, measuredAt, entities.WearableSourceWhoop)
		measurements = appendMeasurement(measurements, row, colSleep, entities.MeasurementTypeSleepDuration, clientID, recordedBy, measuredAt, entities.WearableSourceWhoop)
		measurements = appendMeasurement(measurements, row, colRecovery, entities.MeasurementTypeRecovery, clientID, recordedBy, measuredAt, entities.WearableSourceWhoop)
		measurements = appendMeasurement(measurements, row, colStrain, entities.MeasurementTypeStrain, clientID, recordedBy, measuredAt, entities.WearableSourceWhoop)
		measurements = appendMeasurement(measurements, row, colRestingHR, entities.MeasurementTypeRestingHR, clientID, recordedBy, measuredAt, entities.WearableSourceWhoop)
	}

	if len(measurements) == 0 {
		return nil, fmt.Errorf("no valid measurements found in WHOOP export")
	}

	return measurements, nil
}

// appendMeasurement parses a numeric value from a CSV column and appends a measurement if valid.
func appendMeasurement(measurements []entities.Measurement, row []string, colIdx int, mt entities.MeasurementType, clientID, recordedBy string, measuredAt time.Time, source entities.WearableSource) []entities.Measurement {
	if colIdx < 0 || colIdx >= len(row) {
		return measurements
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(row[colIdx]), 64)
	if err != nil {
		return measurements
	}
	return append(measurements, entities.Measurement{
		ClientID:        clientID,
		RecordedBy:      recordedBy,
		MeasurementType: mt,
		Value:           val,
		Unit:            entities.UnitForMeasurementType(mt),
		MeasuredAt:      measuredAt,
		Source:          source,
		Notes:           fmt.Sprintf("imported from %s", source),
	})
}
