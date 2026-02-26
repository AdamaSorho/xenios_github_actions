package usecase

import (
	"context"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ExtractNutritionUseCase processes a nutrition CSV artifact, computes daily totals
// and rolling averages, and stores the results as measurements and a summary.
type ExtractNutritionUseCase struct {
	artifactRepo    repository.ArtifactRepository
	fileStorage     repository.FileStorageRepository
	measurementRepo repository.MeasurementRepository
	summaryRepo     repository.NutritionSummaryRepository
	auditRepo       repository.AuditRepository
	parser          repository.NutritionParser
}

// NewExtractNutritionUseCase creates a new ExtractNutritionUseCase.
func NewExtractNutritionUseCase(
	artifactRepo repository.ArtifactRepository,
	fileStorage repository.FileStorageRepository,
	measurementRepo repository.MeasurementRepository,
	summaryRepo repository.NutritionSummaryRepository,
	auditRepo repository.AuditRepository,
	parser repository.NutritionParser,
) *ExtractNutritionUseCase {
	return &ExtractNutritionUseCase{
		artifactRepo:    artifactRepo,
		fileStorage:     fileStorage,
		measurementRepo: measurementRepo,
		summaryRepo:     summaryRepo,
		auditRepo:       auditRepo,
		parser:          parser,
	}
}

// ExtractNutritionInput holds the input for the nutrition extraction use case.
type ExtractNutritionInput struct {
	ArtifactID string
	ClientID   string
	CoachID    string
}

// ExtractNutritionOutput holds the result of the nutrition extraction.
type ExtractNutritionOutput struct {
	RowsProcessed int      `json:"rows_processed"`
	RowsSkipped   int      `json:"rows_skipped"`
	DaysProcessed int      `json:"days_processed"`
	Format        string   `json:"format"`
	Errors        []string `json:"errors,omitempty"`
}

// Execute downloads the CSV from storage, parses it, computes daily totals and
// averages, and stores the results.
func (uc *ExtractNutritionUseCase) Execute(ctx context.Context, input ExtractNutritionInput) (*ExtractNutritionOutput, error) {
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	artifact, err := uc.fetchArtifact(ctx, input.ArtifactID)
	if err != nil {
		return nil, err
	}

	csvData, err := uc.fileStorage.Download(ctx, artifact.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("download CSV: %w", err)
	}

	parseResult, err := uc.parser.Parse(csvData)
	if err != nil {
		return nil, &ValidationError{Message: fmt.Sprintf("parse CSV: %s", err)}
	}

	dailyTotals := entities.ComputeDailyTotals(parseResult.Rows)

	measurements := buildMeasurements(dailyTotals, input.ClientID, input.CoachID, input.ArtifactID)
	if len(measurements) > 0 {
		if err := uc.measurementRepo.CreateBatch(ctx, measurements); err != nil {
			return nil, fmt.Errorf("store measurements: %w", err)
		}
	}

	summary := buildNutritionSummary(dailyTotals, input.ClientID, input.ArtifactID)
	if _, err := uc.summaryRepo.Upsert(ctx, summary); err != nil {
		return nil, fmt.Errorf("store nutrition summary: %w", err)
	}

	output := &ExtractNutritionOutput{
		RowsProcessed: parseResult.RowsParsed,
		RowsSkipped:   parseResult.RowsSkipped,
		DaysProcessed: len(dailyTotals),
		Format:        parseResult.Format,
		Errors:        parseResult.Errors,
	}

	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "nutrition.csv_imported",
		EntityType: "artifact",
		EntityID:   input.ArtifactID,
		Metadata: map[string]interface{}{
			"rows_processed": output.RowsProcessed,
			"rows_skipped":   output.RowsSkipped,
			"days_processed": output.DaysProcessed,
			"format":         output.Format,
			"client_id":      input.ClientID,
		},
	})

	return output, nil
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

func (uc *ExtractNutritionUseCase) fetchArtifact(ctx context.Context, artifactID string) (*entities.Artifact, error) {
	artifact, err := uc.artifactRepo.FindByID(ctx, artifactID)
	if err != nil {
		return nil, fmt.Errorf("find artifact: %w", err)
	}
	if artifact == nil {
		return nil, &ValidationError{Message: "artifact not found"}
	}
	if artifact.Status != entities.ArtifactStatusUploaded {
		return nil, &ValidationError{Message: fmt.Sprintf("artifact status is %s, expected uploaded", artifact.Status)}
	}
	return artifact, nil
}

func buildMeasurements(totals []entities.NutritionDailyTotal, clientID, coachID, artifactID string) []*entities.Measurement {
	var measurements []*entities.Measurement
	for _, dt := range totals {
		for _, nt := range entities.AllNutritionTypes() {
			var value float64
			switch nt {
			case entities.NutritionTypeCalories:
				value = dt.Calories
			case entities.NutritionTypeProtein:
				value = dt.Protein
			case entities.NutritionTypeCarbs:
				value = dt.Carbs
			case entities.NutritionTypeFat:
				value = dt.Fat
			case entities.NutritionTypeFiber:
				value = dt.Fiber
			}
			measurements = append(measurements, &entities.Measurement{
				ClientID:        clientID,
				RecordedBy:      coachID,
				MeasurementType: string(nt),
				Value:           value,
				Unit:            entities.NutritionUnit[nt],
				MeasuredAt:      dt.Date,
				ArtifactID:      artifactID,
			})
		}
	}
	return measurements
}

func buildNutritionSummary(totals []entities.NutritionDailyTotal, clientID, artifactID string) *entities.NutritionSummary {
	avg7 := entities.ComputeAverages(totals, 7)
	avg14 := entities.ComputeAverages(totals, 14)
	avg30 := entities.ComputeAverages(totals, 30)

	return &entities.NutritionSummary{
		ClientID:       clientID,
		ArtifactID:     artifactID,
		DaysCount:      len(totals),
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
	}
}
