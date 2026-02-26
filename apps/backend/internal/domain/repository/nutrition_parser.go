package repository

import "github.com/xenios/backend/internal/domain/entities"

// NutritionParseResult holds the result of parsing a nutrition CSV.
type NutritionParseResult struct {
	Format      string
	Rows        []entities.NutritionRow
	RowsParsed  int
	RowsSkipped int
	Errors      []string
}

// NutritionParser defines the interface for parsing nutrition CSV data.
type NutritionParser interface {
	// Parse parses raw CSV bytes and returns structured nutrition rows.
	Parse(data []byte) (*NutritionParseResult, error)
}
