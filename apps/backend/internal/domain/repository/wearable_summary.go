package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// WearableSummaryRepository defines operations for persisting wearable summary data.
type WearableSummaryRepository interface {
	// Upsert inserts or updates a wearable summary for a client+source+date combination.
	Upsert(ctx context.Context, summary *entities.WearableSummary) error
}
