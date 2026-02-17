package usecase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
	"github.com/xenios/backend/internal/infrastructure/parser"
)

// ExtractWearableUseCase handles parsing wearable exports and storing normalized measurements.
type ExtractWearableUseCase struct {
	measurementRepo repository.MeasurementRepository
	summaryRepo     repository.WearableSummaryRepository
	auditRepo       repository.AuditRepository
	parsers         []parser.WearableParser
}

// NewExtractWearableUseCase creates a new ExtractWearableUseCase.
func NewExtractWearableUseCase(
	measurementRepo repository.MeasurementRepository,
	summaryRepo repository.WearableSummaryRepository,
	auditRepo repository.AuditRepository,
	parsers []parser.WearableParser,
) *ExtractWearableUseCase {
	return &ExtractWearableUseCase{
		measurementRepo: measurementRepo,
		summaryRepo:     summaryRepo,
		auditRepo:       auditRepo,
		parsers:         parsers,
	}
}

// ExtractWearableInput holds the input data for the use case.
type ExtractWearableInput struct {
	ClientID string
	CoachID  string
	Data     io.Reader
}

// ExtractWearableOutput holds the result of the extraction.
type ExtractWearableOutput struct {
	Source               entities.WearableSource `json:"source"`
	MeasurementsInserted int                     `json:"measurements_inserted"`
	DuplicatesSkipped    int                     `json:"duplicates_skipped"`
}

// Execute parses the wearable export data, inserts measurements, and computes rolling averages.
func (uc *ExtractWearableUseCase) Execute(ctx context.Context, input ExtractWearableInput) (*ExtractWearableOutput, error) {
	if input.ClientID == "" {
		return nil, entities.NewValidationError("client_id is required")
	}
	if input.CoachID == "" {
		return nil, entities.NewValidationError("coach_id is required")
	}
	if input.Data == nil {
		return nil, entities.NewValidationError("data is required")
	}

	// Read all data into memory for format detection and parsing
	rawData, err := io.ReadAll(input.Data)
	if err != nil {
		return nil, fmt.Errorf("read input data: %w", err)
	}
	if len(rawData) == 0 {
		return nil, entities.NewValidationError("empty data")
	}

	// Detect the source format
	selectedParser, err := uc.detectParser(rawData)
	if err != nil {
		return nil, fmt.Errorf("detect format: %w", err)
	}

	// Parse measurements
	measurements, err := selectedParser.Parse(bytes.NewReader(rawData))
	if err != nil {
		return nil, fmt.Errorf("parse %s data: %w", selectedParser.Source(), err)
	}

	// Set client ID on all measurements
	for i := range measurements {
		measurements[i].ClientID = input.ClientID
	}

	// Bulk upsert measurements (skip duplicates)
	inserted, err := uc.measurementRepo.BulkUpsert(ctx, measurements)
	if err != nil {
		return nil, fmt.Errorf("store measurements: %w", err)
	}

	duplicatesSkipped := len(measurements) - inserted

	// Compute and store rolling averages using the latest date from the measurements
	latestDate := findLatestDate(measurements)
	if err := uc.computeRollingAverages(ctx, input.ClientID, selectedParser.Source(), latestDate); err != nil {
		log.Printf("warning: failed to compute rolling averages: %v", err)
	}

	// Log audit event (non-blocking)
	if auditErr := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "wearable.import",
		EntityType: "measurement",
		EntityID:   input.ClientID,
		Metadata: map[string]interface{}{
			"source":       string(selectedParser.Source()),
			"inserted":     inserted,
			"skipped":      duplicatesSkipped,
			"total_parsed": len(measurements),
		},
	}); auditErr != nil {
		log.Printf("audit log error: %v", auditErr)
	}

	return &ExtractWearableOutput{
		Source:               selectedParser.Source(),
		MeasurementsInserted: inserted,
		DuplicatesSkipped:    duplicatesSkipped,
	}, nil
}

func (uc *ExtractWearableUseCase) detectParser(data []byte) (parser.WearableParser, error) {
	for _, p := range uc.parsers {
		if p.DetectFormat(data) {
			return p, nil
		}
	}
	return nil, fmt.Errorf("unrecognized wearable export format")
}

func findLatestDate(measurements []entities.WearableMeasurement) time.Time {
	var latest time.Time
	for _, m := range measurements {
		if m.MeasuredAt.After(latest) {
			latest = m.MeasuredAt
		}
	}
	return latest
}

func (uc *ExtractWearableUseCase) computeRollingAverages(ctx context.Context, clientID string, source entities.WearableSource, referenceDate time.Time) error {
	if referenceDate.IsZero() {
		return nil
	}
	today := time.Date(referenceDate.Year(), referenceDate.Month(), referenceDate.Day(), 0, 0, 0, 0, time.UTC)

	supportedMetrics, ok := entities.SourceMetricSupport[source]
	if !ok {
		return nil
	}

	metrics := make(map[string]interface{})

	for _, mt := range supportedMetrics {
		for _, window := range entities.RollingAverageWindows {
			from := today.AddDate(0, 0, -window)
			avg, err := uc.measurementRepo.GetAverages(ctx, clientID, source, mt, from, today.AddDate(0, 0, 1))
			if err != nil {
				continue // No data for this window
			}

			key := fmt.Sprintf("avg_%s_%dd", mt, window)
			metrics[key] = roundTo2(avg)
		}
	}

	if len(metrics) == 0 {
		return nil
	}

	summary := &entities.WearableSummary{
		ClientID:    clientID,
		Source:      source,
		SummaryDate: today,
		Metrics:     metrics,
	}

	return uc.summaryRepo.Upsert(ctx, summary)
}

func roundTo2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
