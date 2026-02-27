package usecase

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetInsightQueueUseCase retrieves draft insight cards for a coach's approval queue.
type GetInsightQueueUseCase struct {
	insightRepo repository.InsightCardRepository
}

// NewGetInsightQueueUseCase creates a new GetInsightQueueUseCase.
func NewGetInsightQueueUseCase(insightRepo repository.InsightCardRepository) *GetInsightQueueUseCase {
	return &GetInsightQueueUseCase{insightRepo: insightRepo}
}

// InsightQueueInput holds the input for querying the insight queue.
type InsightQueueInput struct {
	CoachID string
	Page    int
	Limit   int
}

// InsightQueueOutput holds the paginated queue results.
type InsightQueueOutput struct {
	Insights   []*entities.InsightCard `json:"insights"`
	Pagination PaginationOutput        `json:"pagination"`
}

// PaginationOutput holds pagination metadata.
type PaginationOutput struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// Execute retrieves draft insights for the coach's approval queue.
func (uc *GetInsightQueueUseCase) Execute(ctx context.Context, input InsightQueueInput) (*InsightQueueOutput, error) {
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	page := input.Page
	if page < 1 {
		page = 1
	}
	limit := input.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	insights, total, err := uc.insightRepo.ListByCoachIDAndStatus(ctx, input.CoachID, entities.InsightStatusDraft, limit, offset)
	if err != nil {
		return nil, err
	}

	if insights == nil {
		insights = []*entities.InsightCard{}
	}

	return &InsightQueueOutput{
		Insights: insights,
		Pagination: PaginationOutput{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}, nil
}
