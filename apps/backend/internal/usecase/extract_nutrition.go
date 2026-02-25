package usecase

import (
	"context"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ExtractNutritionUseCase parses a nutrition CSV artifact and stores daily measurements.
type ExtractNutritionUseCase struct {
	artifactRepo    repository.ArtifactRepository
	fileStorage     repository.FileStorageRepository
	measurementRepo repository.MeasurementRepository
	auditRepo       repository.AuditRepository
	parser          repository.NutritionParser
}

// NewExtractNutritionUseCase creates a new ExtractNutritionUseCase.
func NewExtractNutritionUseCase(
	artifactRepo repository.ArtifactRepository,
	fileStorage repository.FileStorageRepository,
	measurementRepo repository.MeasurementRepository,
	auditRepo repository.AuditRepository,
	parser repository.NutritionParser,
) *ExtractNutritionUseCase {
	return &ExtractNutritionUseCase{
		artifactRepo:    artifactRepo,
		fileStorage:     fileStorage,
		measurementRepo: measurementRepo,
		auditRepo:       auditRepo,
		parser:          parser,
	}
}

// ExtractNutritionInput holds the input for nutrition extraction.
type ExtractNutritionInput struct {
	ArtifactID string
	CoachID    string
}

// ExtractNutritionOutput holds the result of nutrition extraction.
type ExtractNutritionOutput struct {
	ArtifactID  string                     `json:"artifact_id"`
	Format      entities.CSVFormat         `json:"format"`
	TotalDays   int                        `json:"total_days"`
	TotalRows   int                        `json:"total_rows"`
	SkippedRows int                        `json:"skipped_rows"`
	Averages    entities.NutritionAverages `json:"averages"`
}

// Execute downloads the CSV from storage, parses it, stores measurements, and computes averages.
func (uc *ExtractNutritionUseCase) Execute(ctx context.Context, input ExtractNutritionInput) (*ExtractNutritionOutput, error) {
	if input.ArtifactID == "" {
		return nil, &ValidationError{Message: "artifact_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	artifact, err := uc.artifactRepo.FindByID(ctx, input.ArtifactID)
	if err != nil {
		return nil, fmt.Errorf("find artifact: %w", err)
	}
	if artifact == nil {
		return nil, &ValidationError{Message: "artifact not found"}
	}

	if artifact.CoachID != input.CoachID {
		return nil, &AuthenticationError{Message: "not authorized to process this artifact"}
	}

	csvData, err := uc.fileStorage.DownloadObject(ctx, artifact.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("download CSV: %w", err)
	}

	result, err := uc.parser.Parse(csvData)
	if err != nil {
		return nil, &ValidationError{Message: fmt.Sprintf("parse CSV: %s", err.Error())}
	}

	measurements := uc.parser.DailyTotalsToMeasurements(result.DailyTotals, artifact.ClientID, input.CoachID)
	if len(measurements) > 0 {
		if err := uc.measurementRepo.BatchCreate(ctx, measurements); err != nil {
			return nil, fmt.Errorf("store measurements: %w", err)
		}
	}

	averages := uc.parser.ComputeAverages(result.DailyTotals)

	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "nutrition.import",
		EntityType: "artifact",
		EntityID:   input.ArtifactID,
		Metadata: map[string]interface{}{
			"format":       string(result.Format),
			"total_days":   len(result.DailyTotals),
			"total_rows":   result.TotalRows,
			"skipped_rows": result.SkippedRows,
			"client_id":    artifact.ClientID,
		},
	})

	return &ExtractNutritionOutput{
		ArtifactID:  input.ArtifactID,
		Format:      result.Format,
		TotalDays:   len(result.DailyTotals),
		TotalRows:   result.TotalRows,
		SkippedRows: result.SkippedRows,
		Averages:    averages,
	}, nil
}
