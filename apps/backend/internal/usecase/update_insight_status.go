package usecase

import (
	"context"
	"log"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// UpdateInsightStatusUseCase handles updating the status of an insight card.
type UpdateInsightStatusUseCase struct {
	insightRepo repository.InsightCardRepository
	auditRepo   repository.AuditRepository
}

// NewUpdateInsightStatusUseCase creates a new UpdateInsightStatusUseCase.
func NewUpdateInsightStatusUseCase(
	insightRepo repository.InsightCardRepository,
	auditRepo repository.AuditRepository,
) *UpdateInsightStatusUseCase {
	return &UpdateInsightStatusUseCase{
		insightRepo: insightRepo,
		auditRepo:   auditRepo,
	}
}

// Execute updates the status of an insight card and logs an audit event.
func (uc *UpdateInsightStatusUseCase) Execute(ctx context.Context, id string, status entities.InsightStatus) (*entities.InsightCard, error) {
	if id == "" {
		return nil, &ValidationError{Message: "insight_id is required"}
	}
	if !entities.IsValidInsightStatus(status) {
		return nil, &ValidationError{Message: "invalid status"}
	}

	card, err := uc.insightRepo.UpdateStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}

	action := "insight.approve"
	if status == entities.InsightStatusDismissed {
		action = "insight.reject"
	}

	if err := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    card.CoachID,
		Action:     action,
		EntityType: "insight_card",
		EntityID:   card.ID,
	}); err != nil {
		log.Printf("audit log error: %v", err)
	}

	return card, nil
}
