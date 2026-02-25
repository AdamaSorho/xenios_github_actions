package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// EditInsightUseCase handles editing an insight card's title and body.
type EditInsightUseCase struct {
	insightRepo repository.InsightCardRepository
	auditRepo   repository.AuditRepository
}

// NewEditInsightUseCase creates a new EditInsightUseCase.
func NewEditInsightUseCase(
	insightRepo repository.InsightCardRepository,
	auditRepo repository.AuditRepository,
) *EditInsightUseCase {
	return &EditInsightUseCase{
		insightRepo: insightRepo,
		auditRepo:   auditRepo,
	}
}

// EditInsightInput holds the input for editing an insight.
type EditInsightInput struct {
	InsightID string
	CoachID   string
	Title     string
	Body      string
}

// Execute edits an insight card's title and/or body.
func (uc *EditInsightUseCase) Execute(ctx context.Context, input EditInsightInput) (*entities.InsightCard, error) {
	if input.InsightID == "" {
		return nil, &ValidationError{Message: "insight_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if input.Title == "" && input.Body == "" {
		return nil, &ValidationError{Message: "title or body is required"}
	}

	card, err := uc.insightRepo.FindByID(ctx, input.InsightID)
	if err != nil {
		return nil, fmt.Errorf("find insight: %w", err)
	}
	if card == nil {
		return nil, &ValidationError{Message: "insight not found"}
	}

	if card.CoachID != input.CoachID {
		return nil, &AuthorizationError{Message: "not authorized to edit this insight"}
	}

	if card.Status == entities.InsightStatusDismissed || card.Status == entities.InsightStatusShared {
		return nil, &ValidationError{Message: "cannot edit insight in terminal state"}
	}

	metadata := map[string]interface{}{
		"client_id": card.ClientID,
	}

	if input.Title != "" {
		metadata["old_title"] = card.Title
		card.Title = input.Title
	}
	if input.Body != "" {
		metadata["body_changed"] = true
		card.Body = input.Body
	}
	card.UpdatedAt = time.Now()

	if err := uc.insightRepo.Update(ctx, card); err != nil {
		return nil, fmt.Errorf("update insight: %w", err)
	}

	if auditErr := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "insight.edit",
		EntityType: "insight_card",
		EntityID:   card.ID,
		Metadata:   metadata,
	}); auditErr != nil {
		log.Printf("audit log error: %v", auditErr)
	}

	return card, nil
}
