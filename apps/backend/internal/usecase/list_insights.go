package usecase

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ListInsightsUseCase handles listing insight cards by coach and status.
type ListInsightsUseCase struct {
	insightRepo repository.InsightCardRepository
}

// NewListInsightsUseCase creates a new ListInsightsUseCase.
func NewListInsightsUseCase(insightRepo repository.InsightCardRepository) *ListInsightsUseCase {
	return &ListInsightsUseCase{insightRepo: insightRepo}
}

// Execute returns insight cards for a coach filtered by status.
func (uc *ListInsightsUseCase) Execute(ctx context.Context, coachID string, status entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, error) {
	if coachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if !entities.IsValidInsightStatus(status) {
		return nil, &ValidationError{Message: "invalid status"}
	}
	return uc.insightRepo.FindByStatus(ctx, coachID, status, limit, offset)
}
