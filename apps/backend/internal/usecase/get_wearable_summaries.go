package usecase

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetWearableSummariesUseCase handles retrieving wearable summary data.
type GetWearableSummariesUseCase struct {
	wearableRepo    repository.WearableSummaryRepository
	coachClientRepo repository.CoachClientRepository
	auditRepo       repository.AuditRepository
}

// NewGetWearableSummariesUseCase creates a new GetWearableSummariesUseCase.
func NewGetWearableSummariesUseCase(
	wearableRepo repository.WearableSummaryRepository,
	coachClientRepo repository.CoachClientRepository,
	auditRepo repository.AuditRepository,
) *GetWearableSummariesUseCase {
	return &GetWearableSummariesUseCase{
		wearableRepo:    wearableRepo,
		coachClientRepo: coachClientRepo,
		auditRepo:       auditRepo,
	}
}

const defaultWearableDays = 7

// GetWearableSummariesInput holds the input parameters.
type GetWearableSummariesInput struct {
	CoachID  string
	ClientID string
	Days     int
}

// GetWearableSummariesOutput holds the result.
type GetWearableSummariesOutput struct {
	Summaries []*entities.WearableSummary `json:"summaries"`
}

// Execute retrieves wearable summaries for a client.
func (uc *GetWearableSummariesUseCase) Execute(ctx context.Context, input GetWearableSummariesInput) (*GetWearableSummariesOutput, error) {
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}

	// Authorization check
	rel, err := uc.coachClientRepo.FindByCoachAndClient(ctx, input.CoachID, input.ClientID)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, &AuthorizationError{Message: "not authorized to access this client's data"}
	}

	days := input.Days
	if days <= 0 {
		days = defaultWearableDays
	}

	summaries, err := uc.wearableRepo.FindByClientID(ctx, input.ClientID, days)
	if err != nil {
		return nil, err
	}

	if summaries == nil {
		summaries = []*entities.WearableSummary{}
	}

	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "phi.wearable_data_accessed",
		EntityType: "client",
		EntityID:   input.ClientID,
		Metadata: map[string]interface{}{
			"days": days,
		},
	})

	return &GetWearableSummariesOutput{
		Summaries: summaries,
	}, nil
}
