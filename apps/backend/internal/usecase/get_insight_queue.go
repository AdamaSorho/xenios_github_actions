package usecase

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetInsightQueueUseCase retrieves the approval queue of draft insights for a coach.
type GetInsightQueueUseCase struct {
	insightRepo repository.InsightCardRepository
}

// NewGetInsightQueueUseCase creates a new GetInsightQueueUseCase.
func NewGetInsightQueueUseCase(insightRepo repository.InsightCardRepository) *GetInsightQueueUseCase {
	return &GetInsightQueueUseCase{insightRepo: insightRepo}
}

// GetInsightQueueInput holds the input for the queue query.
type GetInsightQueueInput struct {
	CoachID string
	Status  string
	Limit   int
	Offset  int
}

// GetInsightQueueOutput is an alias for the shared InsightListOutput.
type GetInsightQueueOutput = InsightListOutput

// Execute retrieves the insight queue for the given coach.
func (uc *GetInsightQueueUseCase) Execute(ctx context.Context, input GetInsightQueueInput) (*GetInsightQueueOutput, error) {
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	input.Limit = normalizePagination(input.Limit, 100, 20)

	status := input.Status
	if status == "" {
		status = entities.InsightStatusDraft
	}

	insights, total, err := uc.insightRepo.ListByCoach(ctx, repository.InsightCardFilter{
		CoachID: input.CoachID,
		Status:  status,
		Limit:   input.Limit,
		Offset:  input.Offset,
	})
	if err != nil {
		return nil, err
	}

	return newInsightListOutput(insights, total, input.Limit, input.Offset), nil
}
