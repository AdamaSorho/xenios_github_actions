package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/xenios/backend/internal/domain/entities"
)

// GarminParser parses Garmin CSV exports.
type GarminParser struct{}

// NewGarminParser creates a new GarminParser.
func NewGarminParser() *GarminParser {
	return &GarminParser{}
}

func (p *GarminParser) Source() entities.WearableSource {
	return entities.WearableSourceGarmin
}

func (p *GarminParser) DetectFormat(header []byte) bool {
	line := firstLine(header)
	return strings.Contains(line, "Date") &&
		strings.Contains(line, "Steps") &&
		strings.Contains(line, "Stress Score")
}

func (p *GarminParser) Parse(reader io.Reader) ([]entities.WearableMeasurement, error) {
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
			"Resting Heart Rate (bpm)", entities.MeasurementTypeRestingHR, date, p.Source())
	}

	return measurements, nil
}
