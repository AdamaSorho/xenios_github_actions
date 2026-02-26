package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// WearableSummaryRepository defines the interface for managing wearable summaries.
type WearableSummaryRepository interface {
	// Upsert creates or updates a wearable summary for a given client/source/date.
	Upsert(ctx context.Context, summary *entities.WearableSummary) (*entities.WearableSummary, error)

	// FindByClientID retrieves wearable summaries for a client, ordered by date descending.
	FindByClientID(ctx context.Context, clientID string, limit int) ([]*entities.WearableSummary, error)
}
