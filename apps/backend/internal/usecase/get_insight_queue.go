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

// GetInsightQueueOutput holds the output of the queue query.
type GetInsightQueueOutput struct {
	Insights []*entities.InsightCard `json:"insights"`
	Total    int                     `json:"total"`
	Limit    int                     `json:"limit"`
	Offset   int                     `json:"offset"`
}

// Execute retrieves the insight queue for the given coach.
func (uc *GetInsightQueueUseCase) Execute(ctx context.Context, input GetInsightQueueInput) (*GetInsightQueueOutput, error) {
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	status := input.Status
	if status == "" {
		status = entities.InsightStatusDraft
	}

	filter := repository.InsightCardFilter{
		CoachID: input.CoachID,
		Status:  status,
		Limit:   input.Limit,
		Offset:  input.Offset,
	}

	insights, total, err := uc.insightRepo.ListByCoach(ctx, filter)
	if err != nil {
		return nil, err
	}

	if insights == nil {
		insights = []*entities.InsightCard{}
	}

	return &GetInsightQueueOutput{
		Insights: insights,
		Total:    total,
		Limit:    input.Limit,
		Offset:   input.Offset,
	}, nil
}
