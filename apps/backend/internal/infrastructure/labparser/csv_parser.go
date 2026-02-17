package labparser

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xenios/backend/internal/domain/entities"
)

// csvColumns holds the detected column indices for a lab results CSV.
type csvColumns struct {
	name     int
	value    int
	unit     int
	refRange int
}

// ParseCSV parses lab results from CSV data and returns extracted LabResults.
// It auto-detects the column layout by matching header names.
func ParseCSV(data []byte) ([]entities.LabResult, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1 // Allow variable number of fields
	reader.TrimLeadingSpace = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	cols, err := detectCSVColumns(header)
	if err != nil {
		return nil, err
	}

	var results []entities.LabResult

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		result, ok := parseCSVRow(record, cols)
		if ok {
			results = append(results, result)
		}
	}

	return results, nil
}

// detectCSVColumns identifies which columns contain test name, value, unit, and reference range.
func detectCSVColumns(header []string) (csvColumns, error) {
	cols := csvColumns{name: -1, value: -1, unit: -1, refRange: -1}

	for i, h := range header {
		normalized := strings.ToLower(strings.TrimSpace(h))
		switch {
		case isNameColumn(normalized):
			cols.name = i
		case isValueColumn(normalized):
			cols.value = i
		case isUnitColumn(normalized):
			cols.unit = i
		case isRefRangeColumn(normalized):
			cols.refRange = i
		}
	}

	if cols.name < 0 || cols.value < 0 || cols.unit < 0 {
		return cols, fmt.Errorf("CSV missing required columns (need test name, value, and unit); detected: name=%d value=%d unit=%d", cols.name, cols.value, cols.unit)
	}

	return cols, nil
}

func isNameColumn(s string) bool {
	return s == "test name" || s == "test" || s == "analyte" || s == "marker" || s == "component"
}

func isValueColumn(s string) bool {
	return s == "result" || s == "value" || s == "result value"
}

func isUnitColumn(s string) bool {
	return s == "units" || s == "unit" || s == "uom"
}

func isRefRangeColumn(s string) bool {
	return s == "reference range" || s == "ref range" || s == "reference interval" || s == "normal range"
}

// parseCSVRow parses a single CSV row into a LabResult.
// Returns false if the row cannot be parsed (unrecognized marker or unparseable value).
func parseCSVRow(record []string, cols csvColumns) (entities.LabResult, bool) {
	if len(record) <= cols.name || len(record) <= cols.value || len(record) <= cols.unit {
		return entities.LabResult{}, false
	}

	rawName := strings.TrimSpace(record[cols.name])
	if rawName == "" {
		return entities.LabResult{}, false
	}

	normalizedName := strings.ToLower(rawName)
	labType, known := entities.KnownLabMarkers[normalizedName]
	if !known {
		return entities.LabResult{}, false
	}

	valueStr := strings.TrimSpace(record[cols.value])
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return entities.LabResult{}, false
	}

	unit := strings.TrimSpace(record[cols.unit])

	var refLow, refHigh *float64
	if cols.refRange >= 0 && cols.refRange < len(record) {
		refLow, refHigh = parseReferenceRange(strings.TrimSpace(record[cols.refRange]))
	}

	flag := entities.DetermineFlag(value, refLow, refHigh)

	return entities.LabResult{
		MarkerName:    string(labType),
		Value:         value,
		Unit:          unit,
		ReferenceLow:  refLow,
		ReferenceHigh: refHigh,
		Flag:          flag,
		RawName:       rawName,
	}, true
}

// parseReferenceRange parses reference range strings in common formats:
// "70-100", "<200", ">40", "<5.7", "0.4-4.0"
func parseReferenceRange(s string) (low *float64, high *float64) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	// Less than: "<200" or "< 200"
	if strings.HasPrefix(s, "<") {
		val, err := strconv.ParseFloat(strings.TrimSpace(s[1:]), 64)
		if err == nil {
			return nil, &val
		}
		return nil, nil
	}

	// Greater than: ">40" or "> 40"
	if strings.HasPrefix(s, ">") {
		val, err := strconv.ParseFloat(strings.TrimSpace(s[1:]), 64)
		if err == nil {
			return &val, nil
		}
		return nil, nil
	}

	// Range: "70-100" or "0.4-4.0"
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
