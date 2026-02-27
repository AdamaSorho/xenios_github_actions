package repository

import "github.com/xenios/backend/internal/domain/entities"

// NutritionParser defines the interface for parsing nutrition CSV data.
type NutritionParser interface {
	Parse(data []byte) ([]entities.NutritionEntry, error)
}
