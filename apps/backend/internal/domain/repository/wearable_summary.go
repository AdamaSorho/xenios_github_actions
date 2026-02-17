package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// WearableSummaryRepository defines the interface for wearable summary persistence.
type WearableSummaryRepository interface {
	// Upsert creates or updates a wearable summary for a client+source+date.
	Upsert(ctx context.Context, summary *entities.WearableSummary) error
}
