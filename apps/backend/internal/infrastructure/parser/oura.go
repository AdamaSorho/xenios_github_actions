package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/xenios/backend/internal/domain/entities"
)

// OuraParser parses Oura Ring CSV exports.
type OuraParser struct{}

// NewOuraParser creates a new OuraParser.
func NewOuraParser() *OuraParser {
	return &OuraParser{}
}

func (p *OuraParser) Source() entities.WearableSource {
	return entities.WearableSourceOura
}

func (p *OuraParser) DetectFormat(header []byte) bool {
	line := firstLine(header)
	return strings.Contains(line, "HRV Average (ms)") &&
		strings.Contains(line, "Recovery Index")
}

func (p *OuraParser) Parse(reader io.Reader) ([]entities.WearableMeasurement, error) {
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
			"HRV Average (ms)", entities.MeasurementTypeHRV, date, p.Source())
		measurements = appendMetric(measurements, row, colMap,
			"Sleep Duration (hrs)", entities.MeasurementTypeSleepDuration, date, p.Source())
		measurements = appendMetric(measurements, row, colMap,
			"Recovery Index", entities.MeasurementTypeRecovery, date, p.Source())
		measurements = appendMetric(measurements, row, colMap,
			"Steps", entities.MeasurementTypeSteps, date, p.Source())
		measurements = appendMetric(measurements, row, colMap,
			"Resting Heart Rate (bpm)", entities.MeasurementTypeRestingHR, date, p.Source())
	}

	return measurements, nil
}
