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

// WhoopParser parses WHOOP CSV daily-summary exports.
type WhoopParser struct{}

// NewWhoopParser returns a new WhoopParser.
func NewWhoopParser() *WhoopParser { return &WhoopParser{} }

func (p *WhoopParser) Source() entities.WearableSource {
	return entities.WearableSourceWhoop
}

func (p *WhoopParser) DetectFormat(header []byte) bool {
	h := strings.ToLower(string(header))
	return strings.Contains(h, "recovery score") && strings.Contains(h, "strain")
}

func (p *WhoopParser) Parse(reader io.Reader, clientID string) ([]entities.Measurement, error) {
	r := csv.NewReader(reader)
	r.TrimLeadingSpace = true

	headers, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("whoop: read header: %w", err)
	}

	colIdx := mapColumns(headers)

	var measurements []entities.Measurement
	lineNum := 1

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("whoop: read line %d: %w", lineNum+1, err)
		}
		lineNum++

		dateStr := getCol(record, colIdx, "cycle start time")
		if dateStr == "" {
			continue
		}
		measuredAt, err := parseWhoopDate(dateStr)
		if err != nil {
			continue // skip unparseable rows
		}

		addMetric(&measurements, clientID, entities.WearableSourceWhoop, entities.MeasurementTypeHRV,
			getCol(record, colIdx, "heart rate variability (ms)"), measuredAt)

		addMetric(&measurements, clientID, entities.WearableSourceWhoop, entities.MeasurementTypeRestingHR,
			getCol(record, colIdx, "resting heart rate (bpm)"), measuredAt)

		addMetric(&measurements, clientID, entities.WearableSourceWhoop, entities.MeasurementTypeSleepQuality,
			getCol(record, colIdx, "sleep performance %"), measuredAt)

		addMetric(&measurements, clientID, entities.WearableSourceWhoop, entities.MeasurementTypeRecovery,
			getCol(record, colIdx, "recovery score %"), measuredAt)

		addMetric(&measurements, clientID, entities.WearableSourceWhoop, entities.MeasurementTypeStrain,
			getCol(record, colIdx, "strain"), measuredAt)
	}

	if len(measurements) == 0 {
		return nil, fmt.Errorf("whoop: no valid measurements found")
	}
	return measurements, nil
}

func parseWhoopDate(s string) (time.Time, error) {
	// WHOOP exports use "2024-01-01 06:00:00" format
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		t, err = time.Parse("2006-01-02", s)
	}
	return t, err
}

// mapColumns builds a lowercase header-name → index map.
func mapColumns(headers []string) map[string]int {
	m := make(map[string]int, len(headers))
	for i, h := range headers {
		m[strings.ToLower(strings.TrimSpace(h))] = i
	}
	return m
}

// getCol returns the trimmed value at the column identified by lowercase name, or "".
func getCol(record []string, idx map[string]int, name string) string {
	i, ok := idx[name]
	if !ok || i >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[i])
}

// addMetric parses a string value and, if valid, appends a Measurement.
func addMetric(out *[]entities.Measurement, clientID string, src entities.WearableSource, mt entities.MeasurementType, raw string, measuredAt time.Time) {
	if raw == "" {
		return
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return
	}
	*out = append(*out, entities.Measurement{
		ClientID:        clientID,
		Source:          src,
		MeasurementType: mt,
		Value:           v,
		MeasuredAt:      measuredAt,
	})
}
