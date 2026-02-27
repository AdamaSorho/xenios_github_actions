package usecase

import (
	"context"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ExtractNutritionUseCase handles parsing nutrition CSV files and storing measurements.
type ExtractNutritionUseCase struct {
	artifactRepo repository.ArtifactRepository
	fileStorage  repository.FileStorageRepository
	parser       repository.NutritionParser
	measRepo     repository.MeasurementRepository
	summaryRepo  repository.NutritionSummaryRepository
	auditRepo    repository.AuditRepository
}

// NewExtractNutritionUseCase creates a new ExtractNutritionUseCase.
func NewExtractNutritionUseCase(
	artifactRepo repository.ArtifactRepository,
	fileStorage repository.FileStorageRepository,
	parser repository.NutritionParser,
	measRepo repository.MeasurementRepository,
	summaryRepo repository.NutritionSummaryRepository,
	auditRepo repository.AuditRepository,
) *ExtractNutritionUseCase {
	return &ExtractNutritionUseCase{
		artifactRepo: artifactRepo,
		fileStorage:  fileStorage,
		parser:       parser,
		measRepo:     measRepo,
		summaryRepo:  summaryRepo,
		auditRepo:    auditRepo,
	}
}

// ExtractNutritionInput holds the input for processing a nutrition CSV.
type ExtractNutritionInput struct {
	ArtifactID string
	ClientID   string
	CoachID    string
}

// ExtractNutritionOutput holds the result of the nutrition extraction.
type ExtractNutritionOutput struct {
	DaysProcessed int                        `json:"days_processed"`
	EntriesParsed int                        `json:"entries_parsed"`
	Summary       *entities.NutritionSummary `json:"summary"`
}

// Execute downloads the CSV, parses it, stores daily measurements, and computes averages.
func (uc *ExtractNutritionUseCase) Execute(ctx context.Context, input ExtractNutritionInput) (*ExtractNutritionOutput, error) {
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	artifact, err := uc.loadArtifact(ctx, input.ArtifactID)
	if err != nil {
		return nil, err
	}

	csvData, err := uc.fileStorage.GetObjectContent(ctx, artifact.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("download CSV: %w", err)
	}

	entries, err := uc.parser.Parse(csvData)
	if err != nil {
		return nil, fmt.Errorf("parse CSV: %w", err)
	}

	dailyTotals := entities.ComputeDailyTotals(entries)

	measurements := uc.buildMeasurements(dailyTotals, input)
	if err := uc.measRepo.BatchCreate(ctx, measurements); err != nil {
		return nil, fmt.Errorf("store measurements: %w", err)
	}

	summary := uc.buildSummary(dailyTotals, input)
	if err := uc.summaryRepo.Upsert(ctx, summary); err != nil {
		return nil, fmt.Errorf("store summary: %w", err)
	}

	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "nutrition.imported",
		EntityType: "artifact",
		EntityID:   input.ArtifactID,
		Metadata: map[string]interface{}{
			"client_id":      input.ClientID,
			"days_processed": len(dailyTotals),
			"entries_parsed": len(entries),
		},
	})

	return &ExtractNutritionOutput{
		DaysProcessed: len(dailyTotals),
		EntriesParsed: len(entries),
		Summary:       summary,
	}, nil
}

func (uc *ExtractNutritionUseCase) validateInput(input ExtractNutritionInput) error {
	if input.ArtifactID == "" {
		return &ValidationError{Message: "artifact_id is required"}
	}
	if input.ClientID == "" {
		return &ValidationError{Message: "client_id is required"}
	}
	if input.CoachID == "" {
		return &ValidationError{Message: "coach_id is required"}
	}
	return nil
}

func (uc *ExtractNutritionUseCase) loadArtifact(ctx context.Context, artifactID string) (*entities.Artifact, error) {
	artifact, err := uc.artifactRepo.FindByID(ctx, artifactID)
	if err != nil {
		return nil, fmt.Errorf("find artifact: %w", err)
	}
	if artifact == nil {
		return nil, &ValidationError{Message: "artifact not found"}
	}
	return artifact, nil
}

func (uc *ExtractNutritionUseCase) buildMeasurements(dailyTotals []entities.DailyNutrition, input ExtractNutritionInput) []repository.Measurement {
	var measurements []repository.Measurement
	for _, day := range dailyTotals {
		macros := []struct {
			metric string
			unit   string
			value  float64
		}{
			{"calories", "kcal", day.Calories},
			{"protein", "g", day.Protein},
			{"carbs", "g", day.Carbs},
			{"fat", "g", day.Fat},
			{"fiber", "g", day.Fiber},
		}
		for _, m := range macros {
			measurements = append(measurements, repository.Measurement{
				ClientID:        input.ClientID,
				RecordedBy:      input.CoachID,
				MeasurementType: m.metric,
				Value:           m.value,
				Unit:            m.unit,
				MeasuredAt:      day.Date,
				ArtifactID:      input.ArtifactID,
			})
		}
	}
	return measurements
}

func (uc *ExtractNutritionUseCase) buildSummary(dailyTotals []entities.DailyNutrition, input ExtractNutritionInput) *entities.NutritionSummary {
	avg7 := entities.ComputeAverages(dailyTotals, 7)
	avg14 := entities.ComputeAverages(dailyTotals, 14)
	avg30 := entities.ComputeAverages(dailyTotals, 30)

	return &entities.NutritionSummary{
		ClientID:       input.ClientID,
		ArtifactID:     input.ArtifactID,
		AvgCalories7d:  avg7.Calories,
		AvgProtein7d:   avg7.Protein,
		AvgCarbs7d:     avg7.Carbs,
		AvgFat7d:       avg7.Fat,
		AvgFiber7d:     avg7.Fiber,
		AvgCalories14d: avg14.Calories,
		AvgProtein14d:  avg14.Protein,
		AvgCarbs14d:    avg14.Carbs,
		AvgFat14d:      avg14.Fat,
		AvgFiber14d:    avg14.Fiber,
		AvgCalories30d: avg30.Calories,
		AvgProtein30d:  avg30.Protein,
		AvgCarbs30d:    avg30.Carbs,
		AvgFat30d:      avg30.Fat,
		AvgFiber30d:    avg30.Fiber,
		TotalDays:      len(dailyTotals),
	}
}
