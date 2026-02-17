package usecase

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
	"github.com/xenios/backend/internal/infrastructure/nutrition"
)

// ExtractNutritionUseCase handles parsing nutrition CSV files and storing the results.
type ExtractNutritionUseCase struct {
	artifactRepo  repository.ArtifactRepository
	fileStorage   repository.FileStorageRepository
	nutritionRepo repository.NutritionRepository
	auditRepo     repository.AuditRepository
}

// NewExtractNutritionUseCase creates a new ExtractNutritionUseCase.
func NewExtractNutritionUseCase(
	artifactRepo repository.ArtifactRepository,
	fileStorage repository.FileStorageRepository,
	nutritionRepo repository.NutritionRepository,
	auditRepo repository.AuditRepository,
) *ExtractNutritionUseCase {
	return &ExtractNutritionUseCase{
		artifactRepo:  artifactRepo,
		fileStorage:   fileStorage,
		nutritionRepo: nutritionRepo,
		auditRepo:     auditRepo,
	}
}

// ExtractNutritionInput holds the input for the extract nutrition use case.
type ExtractNutritionInput struct {
	ArtifactID string
	CoachID    string
}

// ExtractNutritionOutput holds the result of the nutrition extraction.
type ExtractNutritionOutput struct {
	ArtifactID    string  `json:"artifact_id"`
	Format        string  `json:"format"`
	DaysProcessed int     `json:"days_processed"`
	TotalRows     int     `json:"total_rows"`
	SkippedRows   int     `json:"skipped_rows"`
	AvgCalories7d float64 `json:"avg_calories_7d"`
	AvgProtein7d  float64 `json:"avg_protein_7d"`
}

// Execute downloads the CSV from storage, parses it, stores daily records and summary.
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

	if artifact.Status != entities.ArtifactStatusUploaded {
		return nil, &ValidationError{Message: fmt.Sprintf("artifact status is %s, expected uploaded", artifact.Status)}
	}

	// Download the CSV content from storage
	content, err := uc.fileStorage.DownloadFile(ctx, artifact.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("download file: %w", err)
	}

	// Parse the CSV
	parseResult, err := nutrition.ParseCSV(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("parse csv: %w", err)
	}

	// Convert daily totals to nutrition records
	records := nutrition.ToNutritionRecords(parseResult.DailyTotals, artifact.ClientID, input.CoachID, input.ArtifactID)

	// Store records
	if len(records) > 0 {
		if err := uc.nutritionRepo.SaveRecords(ctx, records); err != nil {
			return nil, fmt.Errorf("save nutrition records: %w", err)
		}
	}

	// Compute averages
	summary := entities.ComputeAverages(parseResult.DailyTotals)
	summary.ClientID = artifact.ClientID
	summary.ArtifactID = input.ArtifactID
	summary.ComputedAt = time.Now()

	// Store summary
	if err := uc.nutritionRepo.UpsertSummary(ctx, &summary); err != nil {
		return nil, fmt.Errorf("upsert nutrition summary: %w", err)
	}

	// Log audit event
	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "nutrition.imported",
		EntityType: "artifact",
		EntityID:   input.ArtifactID,
		Metadata: map[string]interface{}{
			"format":         parseResult.Format.String(),
			"days_processed": len(parseResult.DailyTotals),
			"total_rows":     parseResult.TotalRows,
			"skipped_rows":   parseResult.SkippedRows,
			"client_id":      artifact.ClientID,
		},
	})

	return &ExtractNutritionOutput{
		ArtifactID:    input.ArtifactID,
		Format:        parseResult.Format.String(),
		DaysProcessed: len(parseResult.DailyTotals),
		TotalRows:     parseResult.TotalRows,
		SkippedRows:   parseResult.SkippedRows,
		AvgCalories7d: summary.AvgCalories7d,
		AvgProtein7d:  summary.AvgProtein7d,
	}, nil
}
