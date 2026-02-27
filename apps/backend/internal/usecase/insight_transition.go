package usecase

import (
	"context"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// InsightTransitionUseCase handles status transitions for insight cards.
// It consolidates approve, dismiss, and share operations to avoid code duplication.
type InsightTransitionUseCase struct {
	insightRepo repository.InsightCardRepository
	auditRepo   repository.AuditRepository
}

// NewInsightTransitionUseCase creates a new InsightTransitionUseCase.
func NewInsightTransitionUseCase(
	insightRepo repository.InsightCardRepository,
	auditRepo repository.AuditRepository,
) *InsightTransitionUseCase {
	return &InsightTransitionUseCase{
		insightRepo: insightRepo,
		auditRepo:   auditRepo,
	}
}

// TransitionInput holds the input for a status transition.
type TransitionInput struct {
	InsightID   string
	CoachID     string
	TargetStatus entities.InsightStatus
	AuditAction string
}

// Execute validates and performs a status transition on an insight card.
func (uc *InsightTransitionUseCase) Execute(ctx context.Context, input TransitionInput) (*entities.InsightCard, error) {
	if input.InsightID == "" {
		return nil, &ValidationError{Message: "insight_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	insight, err := uc.insightRepo.FindByID(ctx, input.InsightID)
	if err != nil {
		return nil, fmt.Errorf("find insight: %w", err)
	}
	if insight == nil {
		return nil, &ValidationError{Message: "insight not found"}
	}

	if insight.CoachID != input.CoachID {
		return nil, &AuthenticationError{Message: "not authorized to modify this insight"}
	}

	if err := entities.ValidateTransition(insight.Status, input.TargetStatus); err != nil {
		return nil, &TransitionError{
			Message:    err.Error(),
			FromStatus: string(insight.Status),
			ToStatus:   string(input.TargetStatus),
		}
	}

	updated, err := uc.insightRepo.UpdateStatus(ctx, input.InsightID, input.TargetStatus)
	if err != nil {
		return nil, fmt.Errorf("update insight status: %w", err)
	}

	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     input.AuditAction,
		EntityType: "insight_card",
		EntityID:   input.InsightID,
		Metadata: map[string]interface{}{
			"from_status": string(insight.Status),
			"to_status":   string(input.TargetStatus),
			"client_id":   insight.ClientID,
		},
	})

	return updated, nil
}

// TransitionError represents an invalid status transition.
type TransitionError struct {
	Message    string
	FromStatus string
	ToStatus   string
}

func (e *TransitionError) Error() string {
	return e.Message
}

// IsTransitionError checks whether the given error is a TransitionError.
func IsTransitionError(err error) bool {
	_, ok := err.(*TransitionError)
	return ok
}
