package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/xenios/backend/internal/domain/entities"
)

// AppleHealthParser parses Apple Health CSV exports.
type AppleHealthParser struct{}

// NewAppleHealthParser creates a new AppleHealthParser.
func NewAppleHealthParser() *AppleHealthParser {
	return &AppleHealthParser{}
}

func (p *AppleHealthParser) Source() entities.WearableSource {
	return entities.WearableSourceAppleHealth
}

func (p *AppleHealthParser) DetectFormat(header []byte) bool {
	line := firstLine(header)
	return strings.Contains(line, "Heart Rate Variability (ms)") &&
		strings.Contains(line, "Resting Heart Rate (bpm)")
}

func (p *AppleHealthParser) Parse(reader io.Reader) ([]entities.WearableMeasurement, error) {
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
		date, err := parseDateColumn(row, colMap)
		if err != nil {
			continue
		}

		measurements = appendMetric(measurements, row, colMap,
			"Heart Rate Variability (ms)", entities.MeasurementTypeHRV, date, p.Source())
		measurements = appendMetric(measurements, row, colMap,
			"Sleep Duration (hrs)", entities.MeasurementTypeSleepDuration, date, p.Source())
		measurements = appendMetric(measurements, row, colMap,
			"Steps", entities.MeasurementTypeSteps, date, p.Source())
		measurements = appendMetric(measurements, row, colMap,
			"Resting Heart Rate (bpm)", entities.MeasurementTypeRestingHR, date, p.Source())
	}

	return measurements, nil
}
