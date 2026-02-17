package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// WhoopParser parses WHOOP CSV exports.
type WhoopParser struct{}

// NewWhoopParser creates a new WhoopParser.
func NewWhoopParser() *WhoopParser {
	return &WhoopParser{}
}

func (p *WhoopParser) Source() entities.WearableSource {
	return entities.WearableSourceWhoop
}

func (p *WhoopParser) DetectFormat(header []byte) bool {
	line := firstLine(header)
	return strings.Contains(line, "Cycle Start Time") &&
		strings.Contains(line, "HRV (ms)")
}

func (p *WhoopParser) Parse(reader io.Reader) ([]entities.WearableMeasurement, error) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}
	if len(records) < 1 {
		return nil, fmt.Errorf("empty CSV file")
	}

	headerRow := records[0]
	colMap := buildColumnMap(headerRow)

	var measurements []entities.WearableMeasurement

	for _, row := range records[1:] {
		date, err := parseWhoopDate(row, colMap)
		if err != nil {
			continue
		}

		measurements = appendMetric(measurements, row, colMap,
			"HRV (ms)", entities.MeasurementTypeHRV, date, p.Source())
		measurements = appendMetric(measurements, row, colMap,
			"Sleep Duration (hrs)", entities.MeasurementTypeSleepDuration, date, p.Source())
		measurements = appendMetric(measurements, row, colMap,
			"Recovery Score (%)", entities.MeasurementTypeRecovery, date, p.Source())
		measurements = appendMetric(measurements, row, colMap,
			"Strain Score", entities.MeasurementTypeStrain, date, p.Source())
		measurements = appendMetric(measurements, row, colMap,
			"Resting Heart Rate (bpm)", entities.MeasurementTypeRestingHR, date, p.Source())
	}

	return measurements, nil
}

func parseWhoopDate(row []string, colMap map[string]int) (time.Time, error) {
	if idx, ok := colMap["Cycle Start Time"]; ok && idx < len(row) {
		return parseDateTime(row[idx])
	}
	return time.Time{}, fmt.Errorf("no date column found")
}
