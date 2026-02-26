package usecase

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ApproveInsightUseCase handles approving a draft insight card.
type ApproveInsightUseCase struct {
	InsightActionDeps
}

// NewApproveInsightUseCase creates a new ApproveInsightUseCase.
func NewApproveInsightUseCase(
	insightRepo repository.InsightCardRepository,
	auditRepo repository.AuditRepository,
) *ApproveInsightUseCase {
	return &ApproveInsightUseCase{
		InsightActionDeps: NewInsightActionDeps(insightRepo, auditRepo),
	}
}

// ApproveInsightInput holds the input for approving an insight.
type ApproveInsightInput struct {
	InsightID string
	CoachID   string
}

// Execute approves a draft insight card.
func (uc *ApproveInsightUseCase) Execute(ctx context.Context, input ApproveInsightInput) (*entities.InsightCard, error) {
	card, err := uc.fetchAndAuthorize(ctx, input.InsightID, input.CoachID, "approve")
	if err != nil {
		return nil, err
	}

	if err := card.TransitionTo(entities.InsightStatusApproved); err != nil {
		return nil, err
	}

	return uc.updateAndAudit(ctx, card, "insight.approve", input.CoachID, nil)
}

// AuthorizationError represents a forbidden access attempt.
type AuthorizationError struct {
	Message string
}

func (e *AuthorizationError) Error() string {
	return e.Message
}

// IsAuthorizationError checks whether the given error is an AuthorizationError.
func IsAuthorizationError(err error) bool {
	_, ok := err.(*AuthorizationError)
	return ok
}
