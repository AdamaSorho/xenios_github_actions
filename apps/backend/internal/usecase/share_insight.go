package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ShareInsightUseCase handles sharing an approved insight with a client.
type ShareInsightUseCase struct {
	insightRepo repository.InsightCardRepository
	auditRepo   repository.AuditRepository
}

// NewShareInsightUseCase creates a new ShareInsightUseCase.
func NewShareInsightUseCase(
	insightRepo repository.InsightCardRepository,
	auditRepo repository.AuditRepository,
) *ShareInsightUseCase {
	return &ShareInsightUseCase{
		insightRepo: insightRepo,
		auditRepo:   auditRepo,
	}
}

// ShareInsightInput holds the input for sharing an insight.
type ShareInsightInput struct {
	InsightID string
	CoachID   string
}

// Execute shares an approved insight card with the client.
func (uc *ShareInsightUseCase) Execute(ctx context.Context, input ShareInsightInput) (*entities.InsightCard, error) {
	if input.InsightID == "" {
		return nil, &ValidationError{Message: "insight_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	card, err := uc.insightRepo.FindByID(ctx, input.InsightID)
	if err != nil {
		return nil, fmt.Errorf("find insight: %w", err)
	}
	if card == nil {
		return nil, &ValidationError{Message: "insight not found"}
	}

	if card.CoachID != input.CoachID {
		return nil, &AuthorizationError{Message: "not authorized to share this insight"}
	}

	if err := card.TransitionTo(entities.InsightStatusShared); err != nil {
		return nil, err
	}

	if err := uc.insightRepo.Update(ctx, card); err != nil {
		return nil, fmt.Errorf("update insight: %w", err)
	}

	if auditErr := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "insight.share",
		EntityType: "insight_card",
		EntityID:   card.ID,
		Metadata: map[string]interface{}{
			"client_id": card.ClientID,
			"title":     card.Title,
		},
	}); auditErr != nil {
		log.Printf("audit log error: %v", auditErr)
	}

	return card, nil
}
