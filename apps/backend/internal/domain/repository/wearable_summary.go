package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// WearableSummaryRepository defines the interface for wearable summary data access.
type WearableSummaryRepository interface {
	// Upsert creates or updates a wearable summary for a given client, source, and date.
	Upsert(ctx context.Context, ws *entities.WearableSummary) (*entities.WearableSummary, error)

	// FindByClientID retrieves wearable summaries for a client, ordered by date descending.
	FindByClientID(ctx context.Context, clientID string, limit, offset int) ([]*entities.WearableSummary, error)
}
