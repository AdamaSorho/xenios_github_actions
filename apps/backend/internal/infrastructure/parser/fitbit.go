package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/xenios/backend/internal/domain/entities"
)

// FitbitParser parses Fitbit CSV exports.
type FitbitParser struct{}

// NewFitbitParser creates a new FitbitParser.
func NewFitbitParser() *FitbitParser {
	return &FitbitParser{}
}

func (p *FitbitParser) Source() entities.WearableSource {
	return entities.WearableSourceFitbit
}

func (p *FitbitParser) DetectFormat(header []byte) bool {
	line := firstLine(header)
	return strings.Contains(line, "Date") &&
		strings.Contains(line, "Resting Heart Rate") &&
		strings.Contains(line, "HRV (ms)") &&
		!strings.Contains(line, "Stress Score") &&
		!strings.Contains(line, "Cycle Start Time")
}

func (p *FitbitParser) Parse(reader io.Reader) ([]entities.WearableMeasurement, error) {
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
			"HRV (ms)", entities.MeasurementTypeHRV, date, p.Source())
		measurements = appendMetric(measurements, row, colMap,
			"Sleep Duration (hrs)", entities.MeasurementTypeSleepDuration, date, p.Source())
		measurements = appendMetric(measurements, row, colMap,
			"Steps", entities.MeasurementTypeSteps, date, p.Source())
		measurements = appendMetric(measurements, row, colMap,
			"Resting Heart Rate", entities.MeasurementTypeRestingHR, date, p.Source())
	}

	return measurements, nil
}
