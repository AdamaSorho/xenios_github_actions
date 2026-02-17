package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// NutritionRepository defines the interface for nutrition data persistence.
type NutritionRepository interface {
	// SaveRecords stores a batch of daily nutrition records.
	SaveRecords(ctx context.Context, records []*entities.NutritionRecord) error

	// UpsertSummary creates or updates a nutrition summary for a client/artifact.
	UpsertSummary(ctx context.Context, summary *entities.NutritionSummary) error

	// GetSummaryByClientID retrieves the latest nutrition summary for a client.
	GetSummaryByClientID(ctx context.Context, clientID string) (*entities.NutritionSummary, error)
}
