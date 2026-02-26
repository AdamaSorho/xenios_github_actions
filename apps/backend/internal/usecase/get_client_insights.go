package usecase

import (
	"context"

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

// GetClientInsightsOutput is an alias for the shared InsightListOutput.
type GetClientInsightsOutput = InsightListOutput

// Execute retrieves insights for a specific client.
func (uc *GetClientInsightsUseCase) Execute(ctx context.Context, input GetClientInsightsInput) (*GetClientInsightsOutput, error) {
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	input.Limit = normalizePagination(input.Limit, 100, 20)

	insights, total, err := uc.insightRepo.ListByClient(ctx, repository.InsightCardFilter{
		CoachID:  input.CoachID,
		ClientID: input.ClientID,
		Status:   input.Status,
		Limit:    input.Limit,
		Offset:   input.Offset,
	})
	if err != nil {
		return nil, err
	}

	return newInsightListOutput(insights, total, input.Limit, input.Offset), nil
}
