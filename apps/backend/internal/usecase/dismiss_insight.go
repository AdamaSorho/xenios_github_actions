package usecase

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// DismissInsightUseCase handles dismissing a draft insight card.
type DismissInsightUseCase struct {
	InsightActionDeps
}

// NewDismissInsightUseCase creates a new DismissInsightUseCase.
func NewDismissInsightUseCase(
	insightRepo repository.InsightCardRepository,
	auditRepo repository.AuditRepository,
) *DismissInsightUseCase {
	return &DismissInsightUseCase{
		InsightActionDeps: NewInsightActionDeps(insightRepo, auditRepo),
	}
}

// DismissInsightInput holds the input for dismissing an insight.
type DismissInsightInput struct {
	InsightID string
	CoachID   string
}

// Execute dismisses a draft insight card.
func (uc *DismissInsightUseCase) Execute(ctx context.Context, input DismissInsightInput) (*entities.InsightCard, error) {
	card, err := uc.fetchAndAuthorize(ctx, input.InsightID, input.CoachID, "dismiss")
	if err != nil {
		return nil, err
	}

	if err := card.TransitionTo(entities.InsightStatusDismissed); err != nil {
		return nil, err
	}

	return uc.updateAndAudit(ctx, card, "insight.dismiss", input.CoachID, nil)
}
