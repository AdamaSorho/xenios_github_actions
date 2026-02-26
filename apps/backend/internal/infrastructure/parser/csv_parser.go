package parser

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/xenios/backend/internal/domain/entities"
)

// CSVLabParser parses lab results from CSV files.
type CSVLabParser struct{}

// NewCSVLabParser creates a new CSVLabParser.
func NewCSVLabParser() *CSVLabParser {
	return &CSVLabParser{}
}

// columnMapping tracks which CSV columns map to which fields.
type columnMapping struct {
	name      int
	result    int
	unit      int
	refRange  int
	hasHeader bool
}

// Parse extracts lab markers from CSV content.
func (p *CSVLabParser) Parse(content []byte) ([]entities.ParsedMarker, error) {
	reader := csv.NewReader(bytes.NewReader(content))
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // allow variable field count

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse csv: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("parse csv: empty file")
	}

	mapping := detectColumns(records[0])
	startRow := 0
	if mapping.hasHeader {
		startRow = 1
	}

	var markers []entities.ParsedMarker
	for i := startRow; i < len(records); i++ {
		marker, err := parseRow(records[i], mapping)
		if err != nil {
			continue // skip unparseable rows
		}
		markers = append(markers, *marker)
	}

	if len(markers) == 0 {
		return nil, fmt.Errorf("parse csv: no valid markers found")
	}

	return markers, nil
}

// detectColumns identifies column positions from a header row.
func detectColumns(header []string) columnMapping {
	m := columnMapping{name: -1, result: -1, unit: -1, refRange: -1}

	for i, col := range header {
		lower := strings.ToLower(strings.TrimSpace(col))
		switch {
		case containsAny(lower, "test name", "test", "marker", "analyte", "component"):
			m.name = i
		case containsAny(lower, "result", "value"):
			m.result = i
		case containsAny(lower, "units", "unit"):
			m.unit = i
		case containsAny(lower, "reference range", "reference", "ref range", "range", "reference interval"):
			m.refRange = i
		}
	}

	// If we found at least name and result columns, it's a header
	if m.name >= 0 && m.result >= 0 {
		m.hasHeader = true
		return m
	}

	// Fallback: assume positional columns (name, value, unit, reference)
	return columnMapping{
		name:      0,
		result:    1,
		unit:      2,
		refRange:  3,
		hasHeader: false,
	}
}

// containsAny returns true if s matches any of the candidates.
func containsAny(s string, candidates ...string) bool {
	for _, c := range candidates {
		if s == c {
			return true
		}
	}
	return false
}

// parseRow extracts a single marker from a CSV row.
func parseRow(row []string, m columnMapping) (*entities.ParsedMarker, error) {
	if len(row) <= m.name || len(row) <= m.result {
		return nil, fmt.Errorf("row too short")
	}

	name := strings.TrimSpace(row[m.name])
	if name == "" {
		return nil, fmt.Errorf("empty marker name")
	}

	valueStr := strings.TrimSpace(row[m.result])
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid value %q: %w", valueStr, err)
	}

	unit := ""
	if m.unit >= 0 && m.unit < len(row) {
		unit = strings.TrimSpace(row[m.unit])
	}

	var refLow, refHigh *float64
	if m.refRange >= 0 && m.refRange < len(row) {
		refLow, refHigh = entities.ParseReferenceRange(row[m.refRange])
	}

	return &entities.ParsedMarker{
		Name:          name,
		Value:         value,
		Unit:          unit,
		ReferenceLow:  refLow,
		ReferenceHigh: refHigh,
	}, nil
}

