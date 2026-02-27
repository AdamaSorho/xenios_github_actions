package usecase

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetClientInsightsUseCase retrieves insights for a specific client.
type GetClientInsightsUseCase struct {
	insightRepo repository.InsightCardRepository
}

// NewGetClientInsightsUseCase creates a new GetClientInsightsUseCase.
func NewGetClientInsightsUseCase(insightRepo repository.InsightCardRepository) *GetClientInsightsUseCase {
	return &GetClientInsightsUseCase{insightRepo: insightRepo}
}

// ClientInsightsInput holds the input for querying client insights.
type ClientInsightsInput struct {
	ClientID string
	CoachID  string
	Status   *entities.InsightStatus
	Page     int
	Limit    int
}

// Execute retrieves insights for the specified client.
func (uc *GetClientInsightsUseCase) Execute(ctx context.Context, input ClientInsightsInput) (*InsightQueueOutput, error) {
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}
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

	insights, total, err := uc.insightRepo.ListByClientID(ctx, input.ClientID, input.Status, limit, offset)
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
