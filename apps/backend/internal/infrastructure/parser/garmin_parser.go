package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// GarminParser parses Garmin CSV daily-summary exports.
type GarminParser struct{}

// NewGarminParser returns a new GarminParser.
func NewGarminParser() *GarminParser { return &GarminParser{} }

func (p *GarminParser) Source() entities.WearableSource {
	return entities.WearableSourceGarmin
}

func (p *GarminParser) DetectFormat(header []byte) bool {
	h := strings.ToLower(string(header))
	return strings.Contains(h, "steps") && strings.Contains(h, "stress level")
}

func (p *GarminParser) Parse(reader io.Reader, clientID string) ([]entities.Measurement, error) {
	r := csv.NewReader(reader)
	r.TrimLeadingSpace = true

	headers, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("garmin: read header: %w", err)
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
			return nil, fmt.Errorf("garmin: read line %d: %w", lineNum+1, err)
		}
		lineNum++

		dateStr := getCol(record, colIdx, "date")
		if dateStr == "" {
			continue
		}
		measuredAt, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		addMetric(&measurements, clientID, entities.WearableSourceGarmin, entities.MeasurementTypeSteps,
			getCol(record, colIdx, "steps"), measuredAt)

		addMetric(&measurements, clientID, entities.WearableSourceGarmin, entities.MeasurementTypeRestingHR,
			getCol(record, colIdx, "resting heart rate (bpm)"), measuredAt)

		addMetric(&measurements, clientID, entities.WearableSourceGarmin, entities.MeasurementTypeSleepDuration,
			getCol(record, colIdx, "sleep duration (hrs)"), measuredAt)

		addMetric(&measurements, clientID, entities.WearableSourceGarmin, entities.MeasurementTypeHRV,
			getCol(record, colIdx, "heart rate variability (ms)"), measuredAt)
	}

	if len(measurements) == 0 {
		return nil, fmt.Errorf("garmin: no valid measurements found")
	}
	return measurements, nil
}
