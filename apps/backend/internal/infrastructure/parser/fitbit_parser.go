package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// FitbitParser parses Fitbit JSON daily-summary exports.
type FitbitParser struct{}

// NewFitbitParser returns a new FitbitParser.
func NewFitbitParser() *FitbitParser { return &FitbitParser{} }

func (p *FitbitParser) Source() entities.WearableSource {
	return entities.WearableSourceFitbit
}

func (p *FitbitParser) DetectFormat(header []byte) bool {
	h := strings.ToLower(string(header))
	return strings.Contains(h, "hrv_ms") &&
		strings.Contains(h, "steps") &&
		!strings.Contains(h, "recovery_score")
}

type fitbitRecord struct {
	Date             string   `json:"date"`
	HRVMS            *float64 `json:"hrv_ms"`
	RestingHRBPM     *float64 `json:"resting_hr_bpm"`
	SleepDurationHrs *float64 `json:"sleep_duration_hrs"`
	Steps            *float64 `json:"steps"`
}

func (p *FitbitParser) Parse(reader io.Reader, clientID string) ([]entities.Measurement, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("fitbit: read input: %w", err)
	}

	var records []fitbitRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("fitbit: parse JSON: %w", err)
	}

	var measurements []entities.Measurement
	for _, rec := range records {
		measuredAt, err := time.Parse("2006-01-02", rec.Date)
		if err != nil {
			continue
		}

		if rec.HRVMS != nil {
			measurements = append(measurements, entities.Measurement{
				ClientID: clientID, Source: entities.WearableSourceFitbit,
				MeasurementType: entities.MeasurementTypeHRV, Value: *rec.HRVMS, MeasuredAt: measuredAt,
			})
		}
		if rec.RestingHRBPM != nil {
			measurements = append(measurements, entities.Measurement{
				ClientID: clientID, Source: entities.WearableSourceFitbit,
				MeasurementType: entities.MeasurementTypeRestingHR, Value: *rec.RestingHRBPM, MeasuredAt: measuredAt,
			})
		}
		if rec.SleepDurationHrs != nil {
			measurements = append(measurements, entities.Measurement{
				ClientID: clientID, Source: entities.WearableSourceFitbit,
				MeasurementType: entities.MeasurementTypeSleepDuration, Value: *rec.SleepDurationHrs, MeasuredAt: measuredAt,
			})
		}
		if rec.Steps != nil {
			measurements = append(measurements, entities.Measurement{
				ClientID: clientID, Source: entities.WearableSourceFitbit,
				MeasurementType: entities.MeasurementTypeSteps, Value: *rec.Steps, MeasuredAt: measuredAt,
			})
		}
	}

	if len(measurements) == 0 {
		return nil, fmt.Errorf("fitbit: no valid measurements found")
	}
	return measurements, nil
}
