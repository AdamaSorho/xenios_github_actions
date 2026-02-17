package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// WearableSummaryRepository defines the interface for wearable summary data access.
type WearableSummaryRepository interface {
	// Upsert creates or updates a wearable summary for a given client, source, and date.
	Upsert(ctx context.Context, summary *entities.WearableSummary) (*entities.WearableSummary, error)

	// FindByClientID retrieves wearable summaries for a client with filtering.
	FindByClientID(ctx context.Context, filter entities.WearableSummaryFilter) ([]*entities.WearableSummary, int, error)

	// FindAverages computes rolling averages from wearable data for the last N days.
	FindAverages(ctx context.Context, clientID string, days int) (*entities.WearableAverages, error)
}
