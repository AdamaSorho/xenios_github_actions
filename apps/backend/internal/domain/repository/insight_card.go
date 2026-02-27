package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// InsightCardRepository defines the interface for insight card persistence.
type InsightCardRepository interface {
	// Create stores a new insight card.
	Create(ctx context.Context, insight *entities.InsightCard) (*entities.InsightCard, error)

	// FindByID retrieves an insight card by its ID.
	FindByID(ctx context.Context, id string) (*entities.InsightCard, error)

	// ListByCoachIDAndStatus retrieves insights for a coach filtered by status with pagination.
	ListByCoachIDAndStatus(ctx context.Context, coachID string, status entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, int, error)

	// ListByClientID retrieves insights for a specific client with optional status filter and pagination.
	ListByClientID(ctx context.Context, clientID string, status *entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, int, error)

	// UpdateStatus updates the status and related timestamp of an insight card.
	UpdateStatus(ctx context.Context, id string, status entities.InsightStatus) (*entities.InsightCard, error)

	// UpdateContent updates the title and body of an insight card.
	UpdateContent(ctx context.Context, id string, title, body string) (*entities.InsightCard, error)

	// CountByCoachIDAndStatus returns the count of insights for a coach with a given status.
	CountByCoachIDAndStatus(ctx context.Context, coachID string, status entities.InsightStatus) (int, error)
}
