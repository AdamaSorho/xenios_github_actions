package usecase

import (
	"context"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ExtractWearableUseCase processes wearable device exports,
// normalizes measurements, and computes rolling averages.
type ExtractWearableUseCase struct {
	measurementRepo repository.MeasurementRepository
	summaryRepo     repository.WearableSummaryRepository
	auditRepo       repository.AuditRepository
}

// NewExtractWearableUseCase creates a new ExtractWearableUseCase.
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

// ExtractWearableInput contains the input parameters for wearable extraction.
type ExtractWearableInput struct {
	ClientID   string
	CoachID    string
	Source     entities.WearableSource
	Reader     io.Reader
	Parser     repository.WearableParser
	ArtifactID string
}

// ExtractWearableOutput contains the results of the extraction.
type ExtractWearableOutput struct {
	MeasurementsInserted int                    `json:"measurements_inserted"`
	Averages             map[string]interface{} `json:"averages"`
}

// Execute processes a wearable export file, stores measurements, and computes averages.
func (uc *ExtractWearableUseCase) Execute(ctx context.Context, input ExtractWearableInput) (*ExtractWearableOutput, error) {
	if input.ClientID == "" {
		return nil, entities.NewValidationError("client_id is required")
	}
	if input.CoachID == "" {
		return nil, entities.NewValidationError("coach_id is required")
	}
	if !entities.IsValidWearableSource(input.Source) {
		return nil, entities.NewValidationError("invalid wearable source: %s", input.Source)
	}
	if input.Reader == nil {
		return nil, entities.NewValidationError("reader is required")
	}
	if input.Parser == nil {
		return nil, entities.NewValidationError("parser is required")
	}

	measurements, err := input.Parser.Parse(input.Reader, input.ClientID, input.CoachID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse wearable export: %w", err)
	}

	inserted, err := uc.measurementRepo.UpsertBatch(ctx, measurements)
	if err != nil {
		return nil, fmt.Errorf("failed to store measurements: %w", err)
	}

	averages, err := uc.computeAverages(ctx, input.ClientID, input.Source, measurements)
	if err != nil {
		return nil, fmt.Errorf("failed to compute averages: %w", err)
	}

	now := time.Now()
	summary := &entities.WearableSummary{
		ClientID:    input.ClientID,
		Source:      input.Source,
		SummaryDate: now.Truncate(24 * time.Hour),
		Metrics:     averages,
		SyncedAt:    now,
	}
	if err := uc.summaryRepo.Upsert(ctx, summary); err != nil {
		return nil, fmt.Errorf("failed to store wearable summary: %w", err)
	}

	auditEvent := &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "artifact.extract",
		EntityType: "artifact",
		EntityID:   input.ArtifactID,
		Metadata: map[string]interface{}{
			"source":                string(input.Source),
			"measurements_inserted": inserted,
			"client_id":             input.ClientID,
		},
	}
	if err := uc.auditRepo.LogEvent(ctx, auditEvent); err != nil {
		return nil, fmt.Errorf("failed to log audit event: %w", err)
	}

	return &ExtractWearableOutput{
		MeasurementsInserted: inserted,
		Averages:             averages,
	}, nil
}

// computeAverages computes rolling averages for all measurement types found in the data.
func (uc *ExtractWearableUseCase) computeAverages(ctx context.Context, clientID string, source entities.WearableSource, measurements []entities.Measurement) (map[string]interface{}, error) {
	// Collect unique measurement types from parsed data
	typeSet := make(map[entities.MeasurementType]bool)
	for _, m := range measurements {
		typeSet[m.MeasurementType] = true
	}

	averages := make(map[string]interface{})
	now := time.Now()

	for mt := range typeSet {
		for _, window := range entities.RollingAverageWindows {
			since := now.AddDate(0, 0, -window)
			avg, err := uc.measurementRepo.Average(ctx, clientID, mt, source, since)
			if err != nil {
				return nil, fmt.Errorf("failed to compute %d-day average for %s: %w", window, mt, err)
			}
			key := fmt.Sprintf("avg_%s_%dd", mt, window)
			averages[key] = roundToOneDecimal(avg)
		}
	}

	return averages, nil
}

// roundToOneDecimal rounds a float to 1 decimal place.
func roundToOneDecimal(v float64) float64 {
	return math.Round(v*10) / 10
}
