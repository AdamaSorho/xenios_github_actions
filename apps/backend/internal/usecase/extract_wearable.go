package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ExtractWearableUseCase orchestrates parsing wearable data, persisting
// measurements, and computing rolling averages.
type ExtractWearableUseCase struct {
	measurementRepo repository.MeasurementRepository
	summaryRepo     repository.WearableSummaryRepository
	auditRepo       repository.AuditRepository
}

// NewExtractWearableUseCase creates a new use case instance.
func NewExtractWearableUseCase(
	measurementRepo repository.MeasurementRepository,
	summaryRepo repository.WearableSummaryRepository,
	auditRepo repository.AuditRepository,
) *ExtractWearableUseCase {
	return &ExtractWearableUseCase{
		measurementRepo: measurementRepo,
		summaryRepo:     summaryRepo,
		auditRepo:       auditRepo,
	}
}

// ExtractWearableResult holds the outcome of a wearable extraction.
type ExtractWearableResult struct {
	Source          entities.WearableSource `json:"source"`
	TotalParsed     int                    `json:"total_parsed"`
	TotalInserted   int                    `json:"total_inserted"`
	DuplicatesSkipped int                  `json:"duplicates_skipped"`
}

// Execute parses the measurements produced by a parser, stores them, and computes rolling averages.
func (uc *ExtractWearableUseCase) Execute(
	ctx context.Context,
	clientID string,
	measurements []entities.Measurement,
	source entities.WearableSource,
) (*ExtractWearableResult, error) {
	if clientID == "" {
		return nil, entities.NewValidationError("client_id is required")
	}
	if !entities.IsValidWearableSource(source) {
		return nil, entities.NewValidationError("invalid wearable source: %q", source)
	}
	if len(measurements) == 0 {
		return nil, entities.NewValidationError("no measurements to process")
	}

	// Persist measurements (upsert skips duplicates)
	inserted, err := uc.measurementRepo.UpsertBatch(ctx, measurements)
	if err != nil {
		return nil, fmt.Errorf("persist measurements: %w", err)
	}

	// Compute and store rolling averages
	if err := uc.computeAndStoreAverages(ctx, clientID, source); err != nil {
		return nil, fmt.Errorf("compute averages: %w", err)
	}

	// Log audit event (non-blocking - ignore errors)
	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    clientID,
		Action:     "wearable.import",
		EntityType: "wearable_data",
		EntityID:   clientID,
		Metadata: map[string]interface{}{
			"source":         string(source),
			"total_parsed":   len(measurements),
			"total_inserted": inserted,
		},
	})

	return &ExtractWearableResult{
		Source:            source,
		TotalParsed:       len(measurements),
		TotalInserted:     inserted,
		DuplicatesSkipped: len(measurements) - inserted,
	}, nil
}

// computeAndStoreAverages computes 7/14/30-day rolling averages for all metric types
// and upserts the summary.
func (uc *ExtractWearableUseCase) computeAndStoreAverages(ctx context.Context, clientID string, source entities.WearableSource) error {
	now := time.Now()
	windows := []struct {
		days   int
		suffix string
	}{
		{7, "7d"},
		{14, "14d"},
		{30, "30d"},
	}

	type metricConfig struct {
		mt     entities.MeasurementType
		prefix string
	}

	metrics := []metricConfig{
		{entities.MeasurementTypeHRV, "avg_hrv"},
		{entities.MeasurementTypeSleepDuration, "avg_sleep"},
		{entities.MeasurementTypeRecovery, "avg_recovery"},
		{entities.MeasurementTypeRestingHR, "avg_resting_hr"},
		{entities.MeasurementTypeSteps, "avg_steps"},
		{entities.MeasurementTypeStrain, "avg_strain"},
	}

	averages := make(map[string]interface{})

	for _, mc := range metrics {
		for _, w := range windows {
			since := now.AddDate(0, 0, -w.days)
			avg, err := uc.measurementRepo.Average(ctx, clientID, source, mc.mt, since)
			if err != nil {
				return fmt.Errorf("average %s_%s: %w", mc.prefix, w.suffix, err)
			}
			if avg != nil {
				key := mc.prefix + "_" + w.suffix
				averages[key] = *avg
			}
		}
	}

	metricsJSON, err := json.Marshal(averages)
	if err != nil {
		return fmt.Errorf("marshal averages: %w", err)
	}

	return uc.summaryRepo.Upsert(ctx, clientID, source, metricsJSON)
}
