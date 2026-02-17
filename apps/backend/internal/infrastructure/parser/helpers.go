package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// allParsers returns all registered wearable parsers.
func allParsers() []WearableParser {
	return []WearableParser{
		NewWhoopParser(),
		NewGarminParser(),
		NewAppleHealthParser(),
		NewOuraParser(),
		NewFitbitParser(),
	}
}

// DetectSource inspects the header bytes and returns the matching parser.
func DetectSource(header []byte) (WearableParser, error) {
	for _, p := range allParsers() {
		if p.DetectFormat(header) {
			return p, nil
		}
	}
	return nil, fmt.Errorf("unrecognized wearable export format")
}

// buildColumnMap maps column header names to their index.
func buildColumnMap(headers []string) map[string]int {
	m := make(map[string]int, len(headers))
	for i, h := range headers {
		m[strings.TrimSpace(h)] = i
	}
	return m
}

// firstLine extracts the first line from the header bytes.
func firstLine(header []byte) string {
	s := string(header)
	if idx := strings.Index(s, "\n"); idx >= 0 {
		s = s[:idx]
	}
	return strings.TrimSpace(s)
}

// parseDate parses a date string in YYYY-MM-DD format.
func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date %q: %w", s, err)
	}
	return t, nil
}

// parseDateTime parses a datetime string, extracting just the date part.
func parseDateTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	// Try full datetime first
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err == nil {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
	}
	// Fall back to date-only
	return parseDate(s)
}

// parseDateColumn extracts the date from the "Date" column.
func parseDateColumn(row []string, colMap map[string]int) (time.Time, error) {
	idx, ok := colMap["Date"]
	if !ok || idx >= len(row) {
		return time.Time{}, fmt.Errorf("no Date column found")
	}
	return parseDate(row[idx])
}

// parseFloat safely parses a float value from a column.
func parseFloat(row []string, colMap map[string]int, colName string) (float64, bool) {
	idx, ok := colMap[colName]
	if !ok || idx >= len(row) {
		return 0, false
	}
	val := strings.TrimSpace(row[idx])
	if val == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

// appendMetric adds a measurement if the column exists and has a valid value.
func appendMetric(
	measurements []entities.WearableMeasurement,
	row []string,
	colMap map[string]int,
	colName string,
	measurementType entities.MeasurementType,
	date time.Time,
	source entities.WearableSource,
) []entities.WearableMeasurement {
	val, ok := parseFloat(row, colMap, colName)
	if !ok {
		return measurements
	}

	return append(measurements, entities.WearableMeasurement{
		Source:          source,
		MeasurementType: measurementType,
		Value:           val,
		Unit:            entities.UnitForMeasurementType(measurementType),
		MeasuredAt:      date,
	})
}
