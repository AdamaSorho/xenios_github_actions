package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// DismissInsightUseCase handles dismissing a draft insight card.
type DismissInsightUseCase struct {
	insightRepo repository.InsightCardRepository
	auditRepo   repository.AuditRepository
}

// NewDismissInsightUseCase creates a new DismissInsightUseCase.
func NewDismissInsightUseCase(
	insightRepo repository.InsightCardRepository,
	auditRepo repository.AuditRepository,
) *DismissInsightUseCase {
	return &DismissInsightUseCase{
		insightRepo: insightRepo,
		auditRepo:   auditRepo,
	}
}

// DismissInsightInput holds the input for dismissing an insight.
type DismissInsightInput struct {
	InsightID string
	CoachID   string
}

// Execute dismisses a draft insight card.
func (uc *DismissInsightUseCase) Execute(ctx context.Context, input DismissInsightInput) (*entities.InsightCard, error) {
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
		return nil, &AuthorizationError{Message: "not authorized to dismiss this insight"}
	}

	if err := card.TransitionTo(entities.InsightStatusDismissed); err != nil {
		return nil, err
	}

	if err := uc.insightRepo.Update(ctx, card); err != nil {
		return nil, fmt.Errorf("update insight: %w", err)
	}

	if auditErr := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "insight.dismiss",
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
