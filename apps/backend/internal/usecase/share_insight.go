package usecase

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ShareInsightUseCase handles sharing an approved insight with a client.
type ShareInsightUseCase struct {
	InsightActionDeps
}

// NewShareInsightUseCase creates a new ShareInsightUseCase.
func NewShareInsightUseCase(
	insightRepo repository.InsightCardRepository,
	auditRepo repository.AuditRepository,
) *ShareInsightUseCase {
	return &ShareInsightUseCase{
		InsightActionDeps: NewInsightActionDeps(insightRepo, auditRepo),
	}
}

// ShareInsightInput holds the input for sharing an insight.
type ShareInsightInput struct {
	InsightID string
	CoachID   string
}

// Execute shares an approved insight card with the client.
func (uc *ShareInsightUseCase) Execute(ctx context.Context, input ShareInsightInput) (*entities.InsightCard, error) {
	card, err := uc.fetchAndAuthorize(ctx, input.InsightID, input.CoachID, "share")
	if err != nil {
		return nil, err
	}

	if err := card.TransitionTo(entities.InsightStatusShared); err != nil {
		return nil, err
	}

	return uc.updateAndAudit(ctx, card, "insight.share", input.CoachID, nil)
}
