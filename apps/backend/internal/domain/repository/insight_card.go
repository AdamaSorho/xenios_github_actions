package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// InsightCardRepository defines the interface for insight card persistence.
type InsightCardRepository interface {
	Create(ctx context.Context, card *entities.InsightCard) (*entities.InsightCard, error)
	FindByID(ctx context.Context, id string) (*entities.InsightCard, error)
	FindByClientID(ctx context.Context, clientID string, limit, offset int) ([]*entities.InsightCard, error)
	FindByStatus(ctx context.Context, coachID string, status entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, error)
	ExistsByEvidence(ctx context.Context, clientID string, measurementID string) (bool, error)
	UpdateStatus(ctx context.Context, id string, status entities.InsightStatus) (*entities.InsightCard, error)
}
