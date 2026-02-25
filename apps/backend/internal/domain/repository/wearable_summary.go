package repository

import (
	"context"
	"encoding/json"

	"github.com/xenios/backend/internal/domain/entities"
)

// WearableSummaryRepository defines persistence operations for rolling average summaries.
type WearableSummaryRepository interface {
	// Upsert creates or updates the summary for a client + source combination.
	Upsert(ctx context.Context, clientID string, source entities.WearableSource, metrics json.RawMessage) error
}
