package repository

import (
	"context"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// WearableSummaryRepository defines the interface for wearable summary data access.
type WearableSummaryRepository interface {
	// FindByClientIDAndDateRange returns wearable summaries for a client within a date range.
	FindByClientIDAndDateRange(ctx context.Context, clientID string, from, to time.Time) ([]*entities.WearableSummary, error)
}
