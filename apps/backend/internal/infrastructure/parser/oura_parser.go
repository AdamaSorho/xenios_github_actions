package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// OuraParser parses Oura JSON daily-summary exports.
type OuraParser struct{}

// NewOuraParser returns a new OuraParser.
func NewOuraParser() *OuraParser { return &OuraParser{} }

func (p *OuraParser) Source() entities.WearableSource {
	return entities.WearableSourceOura
}

func (p *OuraParser) DetectFormat(header []byte) bool {
	h := strings.ToLower(string(header))
	return strings.Contains(h, "recovery_score") && strings.Contains(h, "hrv_ms")
}

type ouraRecord struct {
	Date             string   `json:"date"`
	HRVMS            *float64 `json:"hrv_ms"`
	RestingHRBPM     *float64 `json:"resting_hr_bpm"`
	SleepDurationHrs *float64 `json:"sleep_duration_hrs"`
	RecoveryScore    *float64 `json:"recovery_score"`
	Steps            *float64 `json:"steps"`
}

func (p *OuraParser) Parse(reader io.Reader, clientID string) ([]entities.Measurement, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("oura: read input: %w", err)
	}

	var records []ouraRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("oura: parse JSON: %w", err)
	}

	var measurements []entities.Measurement
	for _, rec := range records {
		measuredAt, err := time.Parse("2006-01-02", rec.Date)
		if err != nil {
			continue
		}

		if rec.HRVMS != nil {
			measurements = append(measurements, entities.Measurement{
				ClientID: clientID, Source: entities.WearableSourceOura,
				MeasurementType: entities.MeasurementTypeHRV, Value: *rec.HRVMS, MeasuredAt: measuredAt,
			})
		}
		if rec.RestingHRBPM != nil {
			measurements = append(measurements, entities.Measurement{
				ClientID: clientID, Source: entities.WearableSourceOura,
				MeasurementType: entities.MeasurementTypeRestingHR, Value: *rec.RestingHRBPM, MeasuredAt: measuredAt,
			})
		}
		if rec.SleepDurationHrs != nil {
			measurements = append(measurements, entities.Measurement{
				ClientID: clientID, Source: entities.WearableSourceOura,
				MeasurementType: entities.MeasurementTypeSleepDuration, Value: *rec.SleepDurationHrs, MeasuredAt: measuredAt,
			})
		}
		if rec.RecoveryScore != nil {
			measurements = append(measurements, entities.Measurement{
				ClientID: clientID, Source: entities.WearableSourceOura,
				MeasurementType: entities.MeasurementTypeRecovery, Value: *rec.RecoveryScore, MeasuredAt: measuredAt,
			})
		}
		if rec.Steps != nil {
			measurements = append(measurements, entities.Measurement{
				ClientID: clientID, Source: entities.WearableSourceOura,
				MeasurementType: entities.MeasurementTypeSteps, Value: *rec.Steps, MeasuredAt: measuredAt,
			})
		}
	}

	if len(measurements) == 0 {
		return nil, fmt.Errorf("oura: no valid measurements found")
	}
	return measurements, nil
}
