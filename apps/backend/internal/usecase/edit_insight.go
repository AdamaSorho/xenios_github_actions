package usecase

import (
	"context"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// EditInsightUseCase handles editing an insight card's title and body.
type EditInsightUseCase struct {
	InsightActionDeps
}

// NewEditInsightUseCase creates a new EditInsightUseCase.
func NewEditInsightUseCase(
	insightRepo repository.InsightCardRepository,
	auditRepo repository.AuditRepository,
) *EditInsightUseCase {
	return &EditInsightUseCase{
		InsightActionDeps: NewInsightActionDeps(insightRepo, auditRepo),
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
	if input.Title == "" && input.Body == "" {
		return nil, &ValidationError{Message: "title or body is required"}
	}

	card, err := uc.fetchAndAuthorize(ctx, input.InsightID, input.CoachID, "edit")
	if err != nil {
		return nil, err
	}

	if card.Status == entities.InsightStatusDismissed || card.Status == entities.InsightStatusShared {
		return nil, &ValidationError{Message: "cannot edit insight in terminal state"}
	}

	extraMetadata := map[string]interface{}{}
	if input.Title != "" {
		extraMetadata["old_title"] = card.Title
		card.Title = input.Title
	}
	if input.Body != "" {
		extraMetadata["body_changed"] = true
		card.Body = input.Body
	}
	card.UpdatedAt = time.Now()

	return uc.updateAndAudit(ctx, card, "insight.edit", input.CoachID, extraMetadata)
}
