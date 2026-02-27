package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// NutritionSummaryRepository defines the interface for nutrition summary persistence.
type NutritionSummaryRepository interface {
	Upsert(ctx context.Context, summary *entities.NutritionSummary) error
}
