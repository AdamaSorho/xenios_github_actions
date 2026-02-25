package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// AppleHealthParser parses Apple Health CSV exports.
type AppleHealthParser struct{}

// NewAppleHealthParser returns a new AppleHealthParser.
func NewAppleHealthParser() *AppleHealthParser { return &AppleHealthParser{} }

func (p *AppleHealthParser) Source() entities.WearableSource {
	return entities.WearableSourceAppleHealth
}

func (p *AppleHealthParser) DetectFormat(header []byte) bool {
	h := strings.ToLower(string(header))
	// Apple Health CSV typically has these columns and no Strain/Recovery/Stress
	return strings.Contains(h, "heart rate variability") &&
		strings.Contains(h, "steps") &&
		!strings.Contains(h, "strain") &&
		!strings.Contains(h, "stress level")
}

func (p *AppleHealthParser) Parse(reader io.Reader, clientID string) ([]entities.Measurement, error) {
	r := csv.NewReader(reader)
	r.TrimLeadingSpace = true

	headers, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("apple_health: read header: %w", err)
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
			return nil, fmt.Errorf("apple_health: read line %d: %w", lineNum+1, err)
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

		addMetric(&measurements, clientID, entities.WearableSourceAppleHealth, entities.MeasurementTypeHRV,
			getCol(record, colIdx, "heart rate variability (ms)"), measuredAt)

		addMetric(&measurements, clientID, entities.WearableSourceAppleHealth, entities.MeasurementTypeRestingHR,
			getCol(record, colIdx, "resting heart rate (bpm)"), measuredAt)

		addMetric(&measurements, clientID, entities.WearableSourceAppleHealth, entities.MeasurementTypeSleepDuration,
			getCol(record, colIdx, "sleep duration (hrs)"), measuredAt)

		addMetric(&measurements, clientID, entities.WearableSourceAppleHealth, entities.MeasurementTypeSteps,
			getCol(record, colIdx, "steps"), measuredAt)
	}

	if len(measurements) == 0 {
		return nil, fmt.Errorf("apple_health: no valid measurements found")
	}
	return measurements, nil
}
