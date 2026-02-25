package repository

import (
	"github.com/xenios/backend/internal/domain/entities"
)

// NutritionParseResult holds the output of parsing a nutrition CSV file.
type NutritionParseResult struct {
	Format      entities.CSVFormat        `json:"format"`
	DailyTotals []entities.DailyNutrition `json:"daily_totals"`
	SkippedRows int                       `json:"skipped_rows"`
	TotalRows   int                       `json:"total_rows"`
}

// NutritionParser defines the interface for parsing nutrition CSV data.
type NutritionParser interface {
	// Parse reads CSV data and returns parsed daily nutrition totals.
	Parse(data []byte) (*NutritionParseResult, error)

	// ComputeAverages calculates rolling averages for the given daily totals.
	ComputeAverages(dailyTotals []entities.DailyNutrition) entities.NutritionAverages

	// DailyTotalsToMeasurements converts daily nutrition totals into Measurement entities.
	DailyTotalsToMeasurements(dailyTotals []entities.DailyNutrition, clientID, recordedBy string) []*entities.Measurement
}
