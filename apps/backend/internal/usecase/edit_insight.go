package usecase

import (
	"context"
	"fmt"

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

// EditInsightInput holds the input for editing an insight card.
type EditInsightInput struct {
	InsightID string
	CoachID   string
	Title     string
	Body      string
}

// Execute validates and edits the insight card's title and body.
func (uc *EditInsightUseCase) Execute(ctx context.Context, input EditInsightInput) (*entities.InsightCard, error) {
	if input.InsightID == "" {
		return nil, &ValidationError{Message: "insight_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if input.Title == "" {
		return nil, &ValidationError{Message: "title is required"}
	}
	if input.Body == "" {
		return nil, &ValidationError{Message: "body is required"}
	}

	insight, err := uc.insightRepo.FindByID(ctx, input.InsightID)
	if err != nil {
		return nil, fmt.Errorf("find insight: %w", err)
	}
	if insight == nil {
		return nil, &ValidationError{Message: "insight not found"}
	}

	if insight.CoachID != input.CoachID {
		return nil, &AuthenticationError{Message: "not authorized to edit this insight"}
	}

	if insight.Status != entities.InsightStatusDraft {
		return nil, &ValidationError{Message: "only draft insights can be edited"}
	}

	updated, err := uc.insightRepo.UpdateContent(ctx, input.InsightID, input.Title, input.Body)
	if err != nil {
		return nil, fmt.Errorf("update insight content: %w", err)
	}

	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "insight.edited",
		EntityType: "insight_card",
		EntityID:   input.InsightID,
		Metadata: map[string]interface{}{
			"new_title": input.Title,
			"client_id": insight.ClientID,
		},
	})

	return updated, nil
}
