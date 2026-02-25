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

// GetClientInsightsInput holds the input for retrieving client insights.
type GetClientInsightsInput struct {
	CoachID  string
	ClientID string
	Status   string
	Limit    int
	Offset   int
}

// GetClientInsightsOutput holds the output of the client insights query.
type GetClientInsightsOutput struct {
	Insights []*entities.InsightCard `json:"insights"`
	Total    int                     `json:"total"`
	Limit    int                     `json:"limit"`
	Offset   int                     `json:"offset"`
}

// Execute retrieves insights for a specific client.
func (uc *GetClientInsightsUseCase) Execute(ctx context.Context, input GetClientInsightsInput) (*GetClientInsightsOutput, error) {
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	filter := repository.InsightCardFilter{
		CoachID:  input.CoachID,
		ClientID: input.ClientID,
		Status:   input.Status,
		Limit:    input.Limit,
		Offset:   input.Offset,
	}

	insights, total, err := uc.insightRepo.ListByClient(ctx, filter)
	if err != nil {
		return nil, err
	}

	if insights == nil {
		insights = []*entities.InsightCard{}
	}

	return &GetClientInsightsOutput{
		Insights: insights,
		Total:    total,
		Limit:    input.Limit,
		Offset:   input.Offset,
	}, nil
}
