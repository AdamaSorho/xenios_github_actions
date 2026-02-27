package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xenios/backend/internal/domain/entities"
)

// CSVLabParser parses lab results from CSV data.
type CSVLabParser struct{}

// NewCSVLabParser creates a new CSVLabParser.
func NewCSVLabParser() *CSVLabParser {
	return &CSVLabParser{}
}

// headerIndices stores column positions from the CSV header.
type headerIndices struct {
	testName  int
	result    int
	units     int
	refRange  int
}

// Parse reads CSV data and returns extracted lab measurements.
// It expects columns for: test name, result, units, reference range.
// Unknown markers are skipped. Returns an error if the CSV format is invalid.
func (p *CSVLabParser) Parse(r io.Reader) ([]entities.LabMeasurement, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // Allow variable field counts
	reader.LazyQuotes = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV header: %w", err)
	}

	indices, err := findHeaderIndices(header)
	if err != nil {
		return nil, err
	}

	var measurements []entities.LabMeasurement
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read CSV row: %w", err)
		}

		m, ok := parseRow(record, indices)
		if ok {
			measurements = append(measurements, m)
		}
	}

	return measurements, nil
}

// findHeaderIndices maps required columns from the header row.
func findHeaderIndices(header []string) (*headerIndices, error) {
	idx := &headerIndices{testName: -1, result: -1, units: -1, refRange: -1}

	for i, col := range header {
		normalized := strings.ToLower(strings.TrimSpace(col))
		switch {
		case containsAny(normalized, "test name", "test", "marker", "analyte", "component"):
			idx.testName = i
		case containsAny(normalized, "result", "value"):
			idx.result = i
		case containsAny(normalized, "units", "unit"):
			idx.units = i
		case containsAny(normalized, "reference range", "reference", "ref range", "normal range"):
			idx.refRange = i
		}
	}

	if idx.testName < 0 || idx.result < 0 {
		return nil, fmt.Errorf("CSV missing required columns: need at least 'Test Name' and 'Result'")
	}
	return idx, nil
}

// containsAny checks if s matches any of the candidates.
func containsAny(s string, candidates ...string) bool {
	for _, c := range candidates {
		if s == c {
			return true
		}
	}
	return false
}

// parseRow extracts a LabMeasurement from a CSV row.
// Returns false if the marker is unrecognized or the value is unparseable.
func parseRow(record []string, idx *headerIndices) (entities.LabMeasurement, bool) {
	if idx.testName >= len(record) || idx.result >= len(record) {
		return entities.LabMeasurement{}, false
	}

	testName := strings.TrimSpace(record[idx.testName])
	markerType := entities.NormalizeMarkerName(testName)
	if markerType == "" {
		return entities.LabMeasurement{}, false
	}

	valueStr := strings.TrimSpace(record[idx.result])
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return entities.LabMeasurement{}, false
	}
	value = entities.RoundToThreeDecimals(value)

	unit := ""
	if idx.units >= 0 && idx.units < len(record) {
		unit = strings.TrimSpace(record[idx.units])
	}

	var refLow, refHigh *float64
	if idx.refRange >= 0 && idx.refRange < len(record) {
		refLow, refHigh = parseReferenceRange(strings.TrimSpace(record[idx.refRange]))
	}

	flag := entities.DetermineFlag(value, refLow, refHigh)

	return entities.LabMeasurement{
		MeasurementType: markerType,
		Value:           value,
		Unit:            unit,
		ReferenceLow:    refLow,
		ReferenceHigh:   refHigh,
		Flag:            flag,
	}, true
}

// parseReferenceRange parses reference range strings like "70-100", "<200", ">40", "0.4-4.0".
func parseReferenceRange(s string) (low *float64, high *float64) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	// Handle "<X" format (upper bound only)
	if strings.HasPrefix(s, "<") || strings.HasPrefix(s, "≤") {
		val, err := strconv.ParseFloat(strings.TrimSpace(s[1:]), 64)
		if err == nil {
			return nil, &val
		}
		return nil, nil
	}

	// Handle ">X" format (lower bound only)
	if strings.HasPrefix(s, ">") || strings.HasPrefix(s, "≥") {
		trimmed := strings.TrimSpace(s[1:])
		// Handle multi-byte ≥/≤
		if len(s) > 1 && s[1] == '=' {
			trimmed = strings.TrimSpace(s[2:])
		}
		val, err := strconv.ParseFloat(trimmed, 64)
		if err == nil {
			return &val, nil
		}
		return nil, nil
	}

	// Handle "X-Y" range format
	parts := strings.SplitN(s, "-", 2)
	if len(parts) == 2 {
		lowVal, errLow := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		highVal, errHigh := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if errLow == nil && errHigh == nil {
			return &lowVal, &highVal
		}
	}

	return nil, nil
}
