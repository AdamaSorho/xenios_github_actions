package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ApproveInsightUseCase handles approving a draft insight card.
type ApproveInsightUseCase struct {
	insightRepo repository.InsightCardRepository
	auditRepo   repository.AuditRepository
}

// NewApproveInsightUseCase creates a new ApproveInsightUseCase.
func NewApproveInsightUseCase(
	insightRepo repository.InsightCardRepository,
	auditRepo repository.AuditRepository,
) *ApproveInsightUseCase {
	return &ApproveInsightUseCase{
		insightRepo: insightRepo,
		auditRepo:   auditRepo,
	}
}

// ApproveInsightInput holds the input for approving an insight.
type ApproveInsightInput struct {
	InsightID string
	CoachID   string
}

// Execute approves a draft insight card.
func (uc *ApproveInsightUseCase) Execute(ctx context.Context, input ApproveInsightInput) (*entities.InsightCard, error) {
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
		return nil, &AuthorizationError{Message: "not authorized to approve this insight"}
	}

	if err := card.TransitionTo(entities.InsightStatusApproved); err != nil {
		return nil, err
	}

	if err := uc.insightRepo.Update(ctx, card); err != nil {
		return nil, fmt.Errorf("update insight: %w", err)
	}

	if auditErr := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "insight.approve",
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
