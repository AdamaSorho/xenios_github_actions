package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetWearableSummariesUseCase retrieves wearable summary data for a client.
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

// GetWearableSummariesInput holds the input for the use case.
type GetWearableSummariesInput struct {
	CoachID  string
	ClientID string
	Limit    int
}

// Execute retrieves wearable summaries after verifying coach-client authorization.
func (uc *GetWearableSummariesUseCase) Execute(ctx context.Context, input GetWearableSummariesInput) ([]*entities.WearableSummary, error) {
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}

	rel, err := uc.coachClientRepo.FindByCoachAndClient(ctx, input.CoachID, input.ClientID)
	if err != nil {
		return nil, fmt.Errorf("check coach-client relationship: %w", err)
	}
	if rel == nil {
		return nil, &AuthorizationError{Message: "not authorized to access this client's data"}
	}

	if input.Limit < 1 || input.Limit > 90 {
		input.Limit = 30
	}

	results, err := uc.wearableRepo.FindByClientID(ctx, input.ClientID, input.Limit)
	if err != nil {
		return nil, fmt.Errorf("find wearable summaries: %w", err)
	}

	if auditErr := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "phi.access",
		EntityType: "client",
		EntityID:   input.ClientID,
		Metadata:   map[string]interface{}{"resource": "wearable_summaries"},
	}); auditErr != nil {
		log.Printf("audit log error: %v", auditErr)
	}

	if results == nil {
		results = []*entities.WearableSummary{}
	}

	return results, nil
}
