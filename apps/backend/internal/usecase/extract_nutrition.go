package usecase

import (
	"context"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// NutritionCSVParser parses CSV data into daily nutrition logs.
type NutritionCSVParser interface {
	Parse(data []byte) ([]entities.NutritionDailyLog, error)
}

// ExtractNutritionUseCase handles parsing nutrition CSV data and storing measurements.
type ExtractNutritionUseCase struct {
	nutritionRepo repository.NutritionRepository
	auditRepo     repository.AuditRepository
	parser        NutritionCSVParser
}

// NewExtractNutritionUseCase creates a new ExtractNutritionUseCase.
func NewExtractNutritionUseCase(
	nutritionRepo repository.NutritionRepository,
	auditRepo repository.AuditRepository,
	parser NutritionCSVParser,
) *ExtractNutritionUseCase {
	return &ExtractNutritionUseCase{
		nutritionRepo: nutritionRepo,
		auditRepo:     auditRepo,
		parser:        parser,
	}
}

// ExtractNutritionInput holds the input for nutrition extraction.
type ExtractNutritionInput struct {
	CSVData    []byte
	ClientID   string
	CoachID    string
	ArtifactID string
}

// ExtractNutritionOutput holds the result of nutrition extraction.
type ExtractNutritionOutput struct {
	DaysProcessed       int                        `json:"days_processed"`
	MeasurementsCreated int                        `json:"measurements_created"`
	Averages            map[int]*entities.NutritionAverage `json:"averages"`
}

var averagePeriods = []int{7, 14, 30}

// Execute parses CSV data, stores measurements, computes averages, and logs audit event.
func (uc *ExtractNutritionUseCase) Execute(ctx context.Context, input ExtractNutritionInput) (*ExtractNutritionOutput, error) {
	if len(input.CSVData) == 0 {
		return nil, &ValidationError{Message: "csv_data is required"}
	}
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if input.ArtifactID == "" {
		return nil, &ValidationError{Message: "artifact_id is required"}
	}

	dailyLogs, err := uc.parser.Parse(input.CSVData)
	if err != nil {
		return nil, &ValidationError{Message: fmt.Sprintf("parse CSV: %s", err.Error())}
	}

	measurements := entities.DailyLogsToMeasurements(dailyLogs, input.ClientID, input.CoachID, input.ArtifactID)
	if err := uc.nutritionRepo.BatchCreateMeasurements(ctx, measurements); err != nil {
		return nil, fmt.Errorf("store measurements: %w", err)
	}

	averages := entities.ComputeNutritionAverages(dailyLogs, averagePeriods)
	var avgSlice []*entities.NutritionAverage
	for _, avg := range averages {
		avg.ClientID = input.ClientID
		avg.SourceArtifactID = input.ArtifactID
		avgSlice = append(avgSlice, avg)
	}
	if len(avgSlice) > 0 {
		if err := uc.nutritionRepo.StoreAverages(ctx, avgSlice); err != nil {
			return nil, fmt.Errorf("store averages: %w", err)
		}
	}

	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "nutrition.import",
		EntityType: "artifact",
		EntityID:   input.ArtifactID,
		Metadata: map[string]interface{}{
			"client_id":      input.ClientID,
			"days_processed": len(dailyLogs),
			"measurements":   len(measurements),
		},
	})

	return &ExtractNutritionOutput{
		DaysProcessed:       len(dailyLogs),
		MeasurementsCreated: len(measurements),
		Averages:            averages,
	}, nil
}
